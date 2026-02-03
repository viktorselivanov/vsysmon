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

func TestRenderDisk_Render(t *testing.T) {
	// Перехватываем stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Подготавливаем снэпшот с тестовыми данными
	snap := &model.Snapshot{
		DiskTPS: 123.45,
		DiskKBs: 678.90,
	}

	renderer := &RenderDisk{}
	renderer.Render(snap)

	// Закрываем писатель и читаем буфер
	w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read stdout pipe: %v", err)
	}
	os.Stdout = oldStdout

	out := buf.String()

	// Проверяем заголовок
	if !strings.Contains(out, "DISK IO") {
		t.Errorf("expected header 'DISK IO', got: %s", out)
	}

	// Проверяем значения TPS и KB/s
	if !strings.Contains(out, "123.45") {
		t.Errorf("expected TPS value 123.45, got: %s", out)
	}
	if !strings.Contains(out, "678.90") {
		t.Errorf("expected KB/s value 678.90, got: %s", out)
	}
}
