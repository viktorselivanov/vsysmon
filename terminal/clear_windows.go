//go:build windows

package terminal

import (
	"os"
	"os/exec"
)

func Clear() {
	cmd := exec.Command("cmd", "/c", "cls") // отчиска экрана для windows
	cmd.Stdout = os.Stdout
	cmd.Run()
}
