//go:build linux
// +build linux

package terminal

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	model "vsysmon/model"
)

func TestRenderListen_Render(t *testing.T) {
	// перехватываем stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
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

	// --- заголовок ---
	if !strings.Contains(out, "LISTENING SOCKETS") {
		t.Fatalf("missing header, got:\n%s", out)
	}

	// --- шапка таблицы ---
	if !strings.Contains(out, "PORT") || !strings.Contains(out, "PID") ||
		!strings.Contains(out, "USER") || !strings.Contains(out, "CMD") {
		t.Fatalf("missing table header, got:\n%s", out)
	}

	// --- первая запись ---
	if !strings.Contains(out, "TCP") ||
		!strings.Contains(out, "80") ||
		!strings.Contains(out, "1234") ||
		!strings.Contains(out, "root") ||
		!strings.Contains(out, "nginx") {
		t.Errorf("missing first listen entry, got:\n%s", out)
	}

	// --- вторая запись ---
	if !strings.Contains(out, "UDP") ||
		!strings.Contains(out, "53") ||
		!strings.Contains(out, "2222") ||
		!strings.Contains(out, "dns") ||
		!strings.Contains(out, "named") {
		t.Errorf("missing second listen entry, got:\n%s", out)
	}
}
