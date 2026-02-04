//go:build linux
// +build linux

package terminal

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	model "vsysmon/internal/model"
)

func TestRenderFS_Render(t *testing.T) {
	// перехватываем stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
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

	checkContains(t, out, "FILESYSTEMS", "missing header")
	checkContains(t, out, "USED(MB)", "missing table header")
	checkContains(t, out, "INODE", "missing table header")

	// проверяем FS записи
	expected := []model.FSStat{
		{Filesystem: "/dev/sda1", MountPoint: "/", UsedMB: 10240, UsedPerc: 50.5, UsedInode: 12345, InodePerc: 12.3},
		{Filesystem: "/dev/sdb1", MountPoint: "/data", UsedMB: 20480, UsedPerc: 75.0, UsedInode: 999, InodePerc: 1.5},
	}

	for _, fs := range expected {
		checkContains(t, out, fs.Filesystem, "missing FS entry")
		checkContains(t, out, fs.MountPoint, "missing FS entry")
		checkContains(t, out, fmt.Sprintf("%v", fs.UsedMB), "missing FS entry")
		checkContains(t, out, fmt.Sprintf("%.1f", fs.UsedPerc), "missing FS entry")
		checkContains(t, out, fmt.Sprintf("%v", fs.UsedInode), "missing FS entry")
		checkContains(t, out, fmt.Sprintf("%.1f", fs.InodePerc), "missing FS entry")
	}
}

// вспомогательная функция для проверки наличия подстроки
func checkContains(t *testing.T, out, substr, msg string) {
	t.Helper()
	if !strings.Contains(out, substr) {
		t.Errorf("%s: expected %q in output", msg, substr)
	}
}
