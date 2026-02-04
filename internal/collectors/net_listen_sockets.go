//go:build linux
// +build linux

package collectors

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	model "vsysmon/internal/model"
)

var (
	readlinkFn   = os.Readlink
	globFn       = filepath.Glob
	readCmdFn    = readCmd
	readUserFn   = readUser
	collectNetFn = collectProcNet
)

type ListenSocketCollector struct{}

func (c *ListenSocketCollector) Name() string { return "listen" }

func (c *ListenSocketCollector) Collect(s *model.Sample) {
	s.Listen = CollectListeningSockets()
}

func CollectListeningSockets() []model.ListenSocket {
	inodeMap := make(map[string]model.ListenSocket) // собираем сокеты в мапу inode для дальнейшего извлечения информации

	collectNetFn("/proc/net/tcp", "TCP", true, inodeMap)
	collectNetFn("/proc/net/tcp6", "TCP6", true, inodeMap)
	collectNetFn("/proc/net/udp", "UDP", false, inodeMap)   // для UDP нет состояния LISTEN
	collectNetFn("/proc/net/udp6", "UDP6", false, inodeMap) // для UDP нет состояния LISTEN

	result := make([]model.ListenSocket, 0, len(inodeMap))

	fds, err := globFn("/proc/[0-9]*/fd/*") // поиск процессов, держащих сокеты/файлы
	if err != nil {
		return nil
	}

	for _, fdPath := range fds {
		link, err := readlinkFn(fdPath)
		if err != nil {
			continue
		}

		if !strings.HasPrefix(link, "socket:[") {
			continue
		}

		inode := strings.TrimSuffix(strings.TrimPrefix(link, "socket:["), "]")

		info, ok := inodeMap[inode] // связываем сокет с процессом
		if !ok {
			continue
		}

		pid := extractPID(fdPath) // извлекаем PID процесса
		if pid <= 0 || pid > int(^uint32(0)) {
			continue
		}

		info.PID = uint32(pid)
		info.Command = readCmdFn(pid)
		info.User = readUserFn(pid)

		result = append(result, info)
		delete(inodeMap, inode)
	}

	return result
}

func collectProcNet(path, proto string, onlyListen bool, inodeMap map[string]model.ListenSocket) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f) // читаем построчно
	first := true

	for sc.Scan() {
		line := sc.Text()
		if first {
			first = false
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		state := fields[3]
		if onlyListen && state != "0A" { // TCP LISTEN
			continue
		}

		// Парсим local_address
		local := fields[1]
		port := parsePort(local)

		// inode — ищем числовое поле после uid
		inode := fields[9]

		// убеждаемся, что это число
		if _, err := strconv.ParseUint(inode, 10, 64); err != nil {
			continue
		}

		inodeMap[inode] = model.ListenSocket{
			Protocol: proto,
			Port:     port,
		}
	}
}

func parsePort(hexAddr string) uint64 { // достаём порт
	p := strings.Split(hexAddr, ":")
	if len(p) != 2 {
		return 0
	}
	port, err := strconv.ParseUint(p[1], 16, 32)
	if err != nil {
		return 0
	}
	return port
}

func extractPID(path string) int { // извлекаем PID
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return 0
	}
	pid, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0 // или другой безопасный дефолт
	}
	return pid
}

func readCmd(pid int) string { // извлекаем имя процесса
	b, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err != nil {
		return "?"
	}
	return strings.TrimSpace(string(b))
}

func readUser(pid int) string { // извлекаем имя пользователя
	fi, err := os.Stat(fmt.Sprintf("/proc/%d", pid))
	if err != nil {
		return "?"
	}

	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return "?" // если тип не тот, безопасно возвращаем значение по умолчанию
	}
	u, err := user.LookupId(strconv.Itoa(int(st.Uid)))
	if err != nil {
		return strconv.Itoa(int(st.Uid))
	}
	return u.Username
}
