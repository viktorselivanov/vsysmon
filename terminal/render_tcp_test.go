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

func TestRenderTCP_Render(t *testing.T) {
	// перехватываем stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	snap := &model.Snapshot{
		TCPStates: map[string]int{
			"ESTABLISHED": 5,
			"LISTEN":      2,
		},
	}

	renderer := &RenderTCP{}
	renderer.Render(snap)

	// закрываем pipe и читаем вывод
	w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read stdout pipe: %v", err)
	}
	os.Stdout = oldStdout

	out := buf.String()

	// заголовок
	if !strings.Contains(out, "TCP STATES") {
		t.Fatalf("missing header, got:\n%s", out)
	}

	if !strings.Contains(out, "STATE") || !strings.Contains(out, "COUNT") {
		t.Fatalf("missing table header, got:\n%s", out)
	}

	// строки со значениями (порядок map не гарантирован)
	if !strings.Contains(out, "ESTABLISHED") || !strings.Contains(out, "5") {
		t.Errorf("expected ESTABLISHED: 5, got:\n%s", out)
	}

	if !strings.Contains(out, "LISTEN") || !strings.Contains(out, "2") {
		t.Errorf("expected LISTEN: 2, got:\n%s", out)
	}
}
