//go:build linux
// +build linux

package collectors

import (
	"log"
	"os"
	"syscall"
	"testing"

	model "vsysmon/internal/model"
)

func TestFSCollector_Collect(t *testing.T) {
	// Сохраняем оригинальные функции, чтобы вернуть после теста
	origOpen := openMounts
	origStatfs := statfs
	defer func() {
		openMounts = origOpen
		statfs = origStatfs
	}()

	// Мокаем /proc/mounts
	mounts := `/dev/sda1 / ext4 rw 0 0
/dev/sdb1 /data ext4 rw 0 0
tmpfs /tmp tmpfs rw 0 0`

	openMounts = func() (*os.File, error) {
		return fakeFile(mounts), nil
	}

	// Мокаем statfs
	statfs = func(path string, st *syscall.Statfs_t) error {
		switch path {
		case "/":
			st.Blocks = 10_000
			st.Bsize = 4096
			st.Bavail = 2_000
			st.Files = 1000
			st.Ffree = 200

		case testUserNginx:
			st.Blocks = 20_000
			st.Bsize = 4096
			st.Bavail = 5_000
			st.Files = 2000
			st.Ffree = 1000

		default:
			return syscall.ENOENT
		}
		return nil
	}

	collector := &FSCollector{}
	s := &model.Sample{}

	collector.Collect(s)

	fs0 := s.FS[0]
	if fs0.MountPoint != "/" {
		t.Errorf("expected mount /, got %s", fs0.MountPoint)
	}

	if fs0.UsedMB == 0 {
		t.Errorf("expected UsedMB > 0")
	}
}

func fakeFile(s string) *os.File {
	r, w, err := os.Pipe()
	if err != nil {
		log.Fatalf("failed to create pipe: %v", err)
	}
	if _, err := w.WriteString(s); err != nil {
		panic(err) // в тестах можно паниковать при ошибке записи
	}
	w.Close()
	return r
}
