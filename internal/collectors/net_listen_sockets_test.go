//go:build linux
// +build linux

package collectors

import (
	"testing"

	model "vsysmon/internal/model"
)

func TestCollectListeningSockets(t *testing.T) {
	// сохраняем оригинальные функции
	origCollect := collectNetFn
	origGlob := globFn
	origReadlink := readlinkFn
	origCmd := readCmdFn
	origUser := readUserFn

	defer func() {
		collectNetFn = origCollect
		globFn = origGlob
		readlinkFn = origReadlink
		readCmdFn = origCmd
		readUserFn = origUser
	}()

	// мокаем сбор из /proc/net
	collectNetFn = func(_, proto string, _ bool, inodeMap map[string]model.ListenSocket) {
		inodeMap["12345"] = model.ListenSocket{
			Protocol: proto,
			Port:     8080,
		}
	}

	// мокаем список fd
	globFn = func(_ string) ([]string, error) {
		return []string{
			"/proc/111/fd/3",
		}, nil
	}

	// мокаем readlink
	readlinkFn = func(_ string) (string, error) {
		return "socket:[12345]", nil
	}

	// мокаем cmd и user
	readCmdFn = func(_ int) string {
		return testCmdNginx
	}

	readUserFn = func(_ int) string {
		return testUserNginx
	}

	// выполняем
	res := CollectListeningSockets()

	if len(res) != 1 {
		t.Fatalf("expected 1 socket, got %d", len(res))
	}

	s := res[0]

	if s.Port != 8080 {
		t.Errorf("expected port 8080, got %d", s.Port)
	}
	if s.Protocol == "" {
		t.Errorf("expected protocol, got empty")
	}
	if s.PID != 111 {
		t.Errorf("expected pid 111, got %d", s.PID)
	}
	if s.Command != "nginx" {
		t.Errorf("expected cmd nginx, got %s", s.Command)
	}
	if s.User != "www-data" {
		t.Errorf("expected user www-data, got %s", s.User)
	}
}
