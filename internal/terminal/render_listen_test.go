//go:build linux
// +build linux

package terminal

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	model "vsysmon/internal/model"
)

func TestRenderListen_Render(t *testing.T) {
	// перехватываем stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	snap := &model.Snapshot{
		Listen: []model.ListenSocket{
			{
				Protocol: "TCP",
				Port:     80,
				PID:      1234,
				User:     "root",
				Command:  "nginx",
			},
			{
				Protocol: "UDP",
				Port:     53,
				PID:      2222,
				User:     "dns",
				Command:  "named",
			},
		},
	}

	renderer := &RenderListen{}
	renderer.Render(snap)

	// закрываем pipe и читаем вывод
	w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read stdout pipe: %v", err)
	}
	os.Stdout = oldStdout

	out := buf.String()

	// заголовок и таблица
	checkContains(t, out, "LISTENING SOCKETS", "missing header")
	checkContains(t, out, "PORT", "missing table header")
	checkContains(t, out, "PID", "missing table header")
	checkContains(t, out, "USER", "missing table header")
	checkContains(t, out, "CMD", "missing table header")

	// проверяем записи
	expected := []model.ListenSocket{
		{Protocol: "TCP", Port: 80, PID: 1234, User: "root", Command: "nginx"},
		{Protocol: "UDP", Port: 53, PID: 2222, User: "dns", Command: "named"},
	}

	for _, l := range expected {
		checkContains(t, out, l.Protocol, "missing listen entry")
		checkContains(t, out, fmt.Sprintf("%d", l.Port), "missing listen entry")
		checkContains(t, out, fmt.Sprintf("%d", l.PID), "missing listen entry")
		checkContains(t, out, l.User, "missing listen entry")
		checkContains(t, out, l.Command, "missing listen entry")
	}
}
