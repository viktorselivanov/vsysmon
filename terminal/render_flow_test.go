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

func TestRenderFlowTop_Render(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	snap := &model.Snapshot{
		FlowTop: []model.FlowTalker{
			{Src: "10.0.0.1:1234", Dst: "10.0.0.2:80", Proto: "TCP", BPS: 1000},
			{Src: "10.0.0.3:4321", Dst: "10.0.0.4:443", Proto: "UDP", BPS: 500},
		},
	}

	renderer := &RenderFlowTop{}
	renderer.Render(snap)

	w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read stdout pipe: %v", err)
	}
	os.Stdout = oldStdout

	out := buf.String()

	// заголовок
	if !strings.Contains(out, "TOP TALKERS — BY FLOW") {
		t.Fatalf("missing header, got:\n%s", out)
	}

	// первая строка
	if !strings.Contains(out, "10.0.0.1:1234") ||
		!strings.Contains(out, "10.0.0.2:80") ||
		!strings.Contains(out, "TCP") ||
		!strings.Contains(out, "1000") {
		t.Errorf("missing first flow entry, got:\n%s", out)
	}

	// вторая строка
	if !strings.Contains(out, "10.0.0.3:4321") ||
		!strings.Contains(out, "10.0.0.4:443") ||
		!strings.Contains(out, "UDP") ||
		!strings.Contains(out, "500") {
		t.Errorf("missing second flow entry, got:\n%s", out)
	}
}
