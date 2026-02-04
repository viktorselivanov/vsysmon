//go:build linux && integration
// +build linux,integration

package collectors_test

import (
	"io"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"vsysmon/collectors"
	"vsysmon/model"
	"vsysmon/ring"
)

func TestFullPipelineFlow(t *testing.T) {
	// --- инициализация кольцевого буфера
	ring.RingInit(5)

	// --- собираем все коллекторы вручную (без конфига)
	collectorsList := []collectors.MetricCollector{
		&collectors.LoadCollector{},
		&collectors.CPUCollector{},
		&collectors.DiskCollector{},
		&collectors.FSCollector{},
		&collectors.TCPStateCollector{},
		&collectors.ListenSocketCollector{},
		&collectors.TopTalkerCollector{},
	}

	done := make(chan struct{})
	defer close(done)

	// --- старт пайплайна
	collectors.StartCollector(done, collectorsList)

	// --- создаём активность, чтобы были данные в снапшотах

	// CPU
	cpuDone := make(chan struct{})
	go func() {
		for i := 0; i < 50_000_000; i++ {
			_ = i * i
		}
		close(cpuDone)
	}()

	// Disk IO
	f, err := os.CreateTemp("", "vsysmon-io")
	if err == nil {
		for i := 0; i < 5000; i++ {
			f.Write(make([]byte, 4096))
		}
		f.Close()
		os.Remove(f.Name())
	}

	// TCP + длительный трафик с несколькими IP
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Logf("failed to start TCP listener: %v", err)
	} else {
		defer ln.Close()

		// Сервер, который отдаёт много данных
		go func() {
			for {
				conn, err := ln.Accept()
				if err != nil {
					return
				}
				go func(c net.Conn) {
					defer c.Close()
					buf := make([]byte, 64*1024)
					end := time.Now().Add(5 * time.Second)
					for time.Now().Before(end) {
						c.Write(buf)
					}
				}(conn)
			}
		}()

		// Генерируем несколько параллельных клиентов с разными локальными IP
		ips := []string{
			"127.0.0.2",
			"127.0.0.3",
			"127.0.0.4",
			"127.0.0.5",
		}

		var wg sync.WaitGroup
		wg.Add(len(ips))

		for _, ip := range ips {
			go func(srcIP string) {
				defer wg.Done()
				d := net.Dialer{
					LocalAddr: &net.TCPAddr{IP: net.ParseIP(srcIP)},
					Timeout:   5 * time.Second,
				}
				conn, err := d.Dial("tcp", ln.Addr().String())
				if err != nil {
					t.Logf("failed to dial from %s: %v", srcIP, err)
					return
				}
				defer conn.Close()

				io.Copy(io.Discard, conn) // читаем поток
			}(ip)
		}

		wg.Wait()
		t.Logf("TCP load completed")
	}

	<-cpuDone

	// ждём появления хотя бы одного сэмпла в кольце
	var snapshots []model.Sample
	timeout := time.After(30 * time.Second)
	tick := time.Tick(100 * time.Millisecond)

	found := false
	for !found {
		select {
		case <-timeout:
			t.Fatal("expected at least one sample in ring, got 0")
		case <-tick:
			snapshots = ring.RingSnapshot()
			if len(snapshots) > 0 {
				found = true
			}
		}
	}

	t.Logf("got %d samples in ring buffer", len(snapshots))

	// проверяем, что статистика изменилась (факт потока)
	snap := snapshots[len(snapshots)-1] // берём последний сэмпл

	if snap.Load <= 0 {
		t.Logf("warning: load average <= 0, might be OK in CI")
	}

	if snap.CPUUser+snap.CPUSys <= 0 {
		t.Logf("warning: CPU activity not detected, might be OK in CI")
	}

	if snap.DiskKBs <= 0 && snap.DiskTPS <= 0 {
		t.Logf("warning: disk activity not detected, might be OK in CI")
	}

	if len(snap.FS) == 0 {
		t.Errorf("expected filesystem stats, got empty")
	}

	if len(snap.TCPStates) == 0 {
		t.Errorf("expected TCP states, got empty")
	}

	if len(snap.Listen) == 0 {
		t.Logf("warning: no listening sockets detected (can be OK in CI)")
	}

	if len(snap.ProtoTop) == 0 {
		t.Errorf("expected proto top talkers, got empty")
	}

	if len(snap.FlowTop) == 0 {
		t.Errorf("expected flow top talkers, got empty")
	}
}
