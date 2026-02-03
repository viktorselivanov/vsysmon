//go:build linux
// +build linux

package collectors

import (
	"bufio"
	"os"
	"strings"
	"syscall"
	model "vsysmon/model"
)

var openMounts = func() (*os.File, error) {
	return os.Open("/proc/mounts")
}

var statfs = func(path string, st *syscall.Statfs_t) error {
	return syscall.Statfs(path, st)
}

type FSCollector struct{}

func (c *FSCollector) Name() string { return "fs" }

func (c *FSCollector) Collect(s *model.Sample) {
	fs, _ := getFSStats()
	s.FS = fs
}

func getFSStats() ([]model.FSStat, error) {
	var stats []model.FSStat

	f, err := openMounts()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		device := fields[0]
		mount := fields[1]

		// только реальные устройства
		if !(strings.HasPrefix(device, "/dev/") || strings.HasPrefix(device, "/dev/mapper/")) {
			continue
		}

		// исключаем snap и /var/snap
		if strings.HasPrefix(mount, "/snap") || strings.HasPrefix(mount, "/var/snap") {
			continue
		}

		var st syscall.Statfs_t

		if err := statfs(mount, &st); err != nil {
			continue
		}

		// df: пропускаем пустые FS
		if st.Blocks == 0 || st.Files == 0 {
			continue
		}

		// расчитываем роцент использования места
		total := st.Blocks * uint64(st.Bsize)
		free := st.Bavail * uint64(st.Bsize)
		used := total - free
		usedPerc := 0.0
		if total > 0 {
			usedPerc = float64(used) / float64(total) * 100
		}
		// расчтьываем количества открытых фалов
		totalInode := st.Files
		freeInode := st.Ffree
		usedInode := totalInode - freeInode
		inodePerc := 0.0
		if totalInode > 0 {
			inodePerc = float64(usedInode) / float64(totalInode) * 100
		}
		// собираем в структуру
		stats = append(stats, model.FSStat{
			Filesystem: device,
			MountPoint: mount,
			UsedMB:     used / 1024 / 1024,
			UsedPerc:   usedPerc,
			UsedInode:  usedInode,
			InodePerc:  inodePerc,
		})
	}

	return stats, nil
}
