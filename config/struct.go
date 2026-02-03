package config

type Config struct {
	CollectLoad       bool `json:"collect_load"`
	CollectCPU        bool `json:"collect_cpu"`
	CollectDisk       bool `json:"collect_disk"`
	CollectFS         bool `json:"collect_fs"`
	CollectTCPStates  bool `json:"collect_tcp_states"`
	CollectListen     bool `json:"collect_listen"`
	CollectTopTalkers bool `json:"collect_top_talkers"`
}

type OSConfig struct {
	Linux   Config `json:"linux"`
	Windows Config `json:"windows"`
}
