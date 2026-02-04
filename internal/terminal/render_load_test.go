//go:build linux
// +build linux

package terminal

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	model "vsysmon/internal/model"
)

func TestLoadRenderer_Render(t *testing.T) {
	// перехватываем stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	snap := &model.Snapshot{
		Load: 1.23,
	}

	renderer := &LoadRenderer{}
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
	if !strings.Contains(out, "LOAD AVERAGE") {
		t.Fatalf("missing header, got:\n%s", out)
	}

	// строка значения
	if !strings.Contains(out, "Load avg: 1.23") {
		t.Errorf("expected load value, got:\n%s", out)
	}
}
