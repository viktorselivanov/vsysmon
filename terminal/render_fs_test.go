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

func TestRenderFS_Render(t *testing.T) {
	// перехватываем stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	snap := &model.Snapshot{
		FS: []model.FSStat{
			{
				Filesystem: "/dev/sda1",
				MountPoint: "/",
				UsedMB:     10240,
				UsedPerc:   50.5,
				UsedInode:  12345,
				InodePerc:  12.3,
			},
			{
				Filesystem: "/dev/sdb1",
				MountPoint: "/data",
				UsedMB:     20480,
				UsedPerc:   75.0,
				UsedInode:  999,
				InodePerc:  1.5,
			},
		},
	}

	renderer := &RenderFS{}
	renderer.Render(snap)

	// закрываем pipe и читаем вывод
	w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read stdout pipe: %v", err)
	}
	os.Stdout = oldStdout

	out := buf.String()

	//заголовки
	if !strings.Contains(out, "FILESYSTEMS") {
		t.Fatalf("missing header, got:\n%s", out)
	}

	if !strings.Contains(out, "USED(MB)") || !strings.Contains(out, "INODE") {
		t.Fatalf("missing table header, got:\n%s", out)
	}

	//первая FS
	if !strings.Contains(out, "/dev/sda1") ||
		!strings.Contains(out, "/") ||
		!strings.Contains(out, "10240") ||
		!strings.Contains(out, "50.5") ||
		!strings.Contains(out, "12345") ||
		!strings.Contains(out, "12.3") {
		t.Errorf("missing first FS entry, got:\n%s", out)
	}

	//вторая FS
	if !strings.Contains(out, "/dev/sdb1") ||
		!strings.Contains(out, "/data") ||
		!strings.Contains(out, "20480") ||
		!strings.Contains(out, "75.0") ||
		!strings.Contains(out, "999") ||
		!strings.Contains(out, "1.5") {
		t.Errorf("missing second FS entry, got:\n%s", out)
	}
}
