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

func TestRenderProtoTop_Render(t *testing.T) {
	// перехватываем stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	snap := &model.Snapshot{
		ProtoTop: []model.ProtoTalker{
			{Proto: "TCP", Bytes: 1000, Perc: 66.6},
			{Proto: "UDP", Bytes: 500, Perc: 33.3},
		},
	}

	renderer := &RenderProtoTop{}
	renderer.Render(snap)

	// закрываем pipe и читаем вывод
	w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read stdout pipe: %v", err)
	}
	os.Stdout = oldStdout

	out := buf.String()

	// заголовки
	if !strings.Contains(out, "TOP TALKERS — BY PROTOCOL") {
		t.Fatalf("missing header, got:\n%s", out)
	}

	if !strings.Contains(out, "PROTO") || !strings.Contains(out, "BYTES/s") {
		t.Fatalf("missing table header, got:\n%s", out)
	}

	// строки данных
	if !strings.Contains(out, "TCP") || !strings.Contains(out, "1000") {
		t.Errorf("expected TCP row, got:\n%s", out)
	}

	if !strings.Contains(out, "UDP") || !strings.Contains(out, "500") {
		t.Errorf("expected UDP row, got:\n%s", out)
	}

	// процент с %
	if !strings.Contains(out, "%") {
		t.Errorf("expected percent sign in output, got:\n%s", out)
	}
}
