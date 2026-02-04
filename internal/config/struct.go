package config

type Config struct {
	CollectLoad       bool `json:"collectLoad"`
	CollectCPU        bool `json:"collectCpu"`
	CollectDisk       bool `json:"collectDisk"`
	CollectFS         bool `json:"collectFs"`
	CollectTCPStates  bool `json:"collectTcpStates"`
	CollectListen     bool `json:"collectListen"`
	CollectTopTalkers bool `json:"collectTopTalkers"`
}

type OSConfig struct {
	Linux   Config `json:"linux"`
	Windows Config `json:"windows"`
}
