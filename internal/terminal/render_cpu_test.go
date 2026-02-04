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

func TestRenderCPU_Render(t *testing.T) {
	// Перехватываем stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	// Подготавливаем снэпшот с тестовыми данными
	snap := &model.Snapshot{
		CPUUser: 12.34,
		CPUSys:  56.78,
		CPUIdle: 30.88,
	}

	renderer := &RenderCPU{}
	renderer.Render(snap)

	// Закрываем рендер и читаем буфер
	w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read stdout pipe: %v", err)
	}
	os.Stdout = oldStdout

	out := buf.String()

	// Проверяем, что вывод содержит заголовок
	if !strings.Contains(out, "CPU USAGE (%)") {
		t.Errorf("expected header 'CPU USAGE (%%)', got: %s", out)
	}

	// Проверяем, что значения CPU присутствуют в выводе
	if !strings.Contains(out, "12.34") {
		t.Errorf("expected User value 12.34, got: %s", out)
	}
	if !strings.Contains(out, "56.78") {
		t.Errorf("expected System value 56.78, got: %s", out)
	}
	if !strings.Contains(out, "30.88") {
		t.Errorf("expected Idle value 30.88, got: %s", out)
	}
}
