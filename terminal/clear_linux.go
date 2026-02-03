//go:build linux

package terminal

import "fmt"

func Clear() {
	fmt.Print("\033[H\033[2J\033[3J") // отчиска экрана для linux
}
