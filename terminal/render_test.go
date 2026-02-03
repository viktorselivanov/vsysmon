//go:build linux
// +build linux

package terminal

import (
	"bytes"
	"io"
	"os"
	"testing"

	model "vsysmon/model"
)

// MockRenderer — мок для MetricRenderer
type MockRenderer struct {
	Called bool
	NameFn string
}

func (m *MockRenderer) Render(s *model.Snapshot) {
	m.Called = true
}

func (m *MockRenderer) Name() string {
	return m.NameFn
}

func TestRender(t *testing.T) {
	// Перехватываем stdout в буфер
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Создаём мок-секции
	m1 := &MockRenderer{NameFn: "mock1"}
	m2 := &MockRenderer{NameFn: "mock2"}

	snap := &model.Snapshot{}

	Render(snap, []MetricRenderer{m1, m2})

	// Закрываем писатель и читаем буфер
	w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read stdout pipe: %v", err)
	}
	os.Stdout = oldStdout

	// Проверяем, что Render вызван у всех секций
	if !m1.Called {
		t.Errorf("expected Render to be called on m1")
	}
	if !m2.Called {
		t.Errorf("expected Render to be called on m2")
	}

	// Проверяем, что заголовок вывода есть
	if !bytes.Contains(buf.Bytes(), []byte("======== SYSTEM SNAPSHOT ========")) {
		t.Errorf("expected header in output")
	}
}
