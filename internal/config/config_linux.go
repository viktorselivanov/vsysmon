//go:build linux
// +build linux

package config

import (
	"encoding/json"
	"os"
)

var oscfg OSConfig // хранит весь конфиг из файла

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path) // читаем конфиг
	if err != nil {
		return Config{}, err
	}

	if err := json.Unmarshal(data, &oscfg); err != nil { // парсим json
		return Config{}, err
	}

	return oscfg.Linux, nil // возращаем только win конфиг
}

func DefaultConfig() Config { // дефолтный конфиг для клиента, либо для сервера при отсутствии файла
	return Config{
		CollectLoad:       true,
		CollectCPU:        true,
		CollectDisk:       true,
		CollectFS:         true,
		CollectTCPStates:  true,
		CollectListen:     true,
		CollectTopTalkers: true,
	}
}
