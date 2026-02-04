//go:build linux && integration
// +build linux,integration

package collectors_test

import (
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"vsysmon/collectors"
	"vsysmon/model"
)

func TestFullPipelineSnapshotFlow(t *testing.T) {
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

	// первый снапшот (инициализация prev-состояний)
	s1 := &model.Sample{}
	for _, c := range collectorsList {
		c.Collect(s1)
	}

	// создаём активность

	// CPU
	done := make(chan struct{})
	go func() {
		for i := 0; i < 50_000_000; i++ {
			_ = i * i
		}
		close(done)
	}()

	// Disk IO
	f, err := os.CreateTemp("", "vsysmon-io")
	if err == nil {
		for i := 0; i < 5000; i++ {
			f.Write(make([]byte, 4096))
		}
		f.Sync() // обязательно сбрасываем
		f.Close()
		os.Remove(f.Name())
	}
	time.Sleep(200 * time.Millisecond) // даём время /proc/diskstats обновиться

	// TCP traffic
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err == nil {
		defer ln.Close()
		go http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(make([]byte, 20_000))
		}))
		http.Get("http://" + ln.Addr().String())
		http.Get("http://" + ln.Addr().String()) // второй запрос, чтобы трафик был виден
	}

	<-done

	time.Sleep(500 * time.Millisecond)

	// второй снапшот
	s2 := &model.Sample{}
	for _, c := range collectorsList {
		c.Collect(s2)
	}

	// проверки

	if s2.Load <= 0 {
		t.Errorf("expected Load > 0, got %.2f", s2.Load)
	}

	if s2.CPUUser+s2.CPUSys <= 0 {
		t.Errorf("expected CPU activity, got user=%.2f sys=%.2f", s2.CPUUser, s2.CPUSys)
	}

	if s2.DiskKBs <= 0 && s2.DiskTPS <= 0 {
		t.Errorf("expected disk activity, got TPS=%.2f KB/s=%.2f", s2.DiskTPS, s2.DiskKBs)
	}

	if len(s2.FS) == 0 {
		t.Errorf("expected filesystem stats, got empty")
	}

	if len(s2.TCPStates) == 0 {
		t.Errorf("expected tcp states, got empty")
	}

	if len(s2.Listen) == 0 {
		t.Logf("warning: no listening sockets detected (can be OK in CI)")
	}

	if len(s2.ProtoTop) == 0 {
		t.Errorf("expected proto top talkers, got empty")
	}

	if len(s2.FlowTop) == 0 {
		t.Errorf("expected flow top talkers, got empty")
	}
}
