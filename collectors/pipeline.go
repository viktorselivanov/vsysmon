package collectors

import (
	"vsysmon/config"
)

func BuildPipeline(cfg config.Config) []MetricCollector {
	var p []MetricCollector

	if cfg.CollectLoad {
		p = append(p, &LoadCollector{})
	}
	if cfg.CollectCPU {
		p = append(p, &CPUCollector{})
	}
	if cfg.CollectDisk {
		p = append(p, &DiskCollector{})
	}
	if cfg.CollectFS {
		p = append(p, &FSCollector{})
	}
	if cfg.CollectTCPStates {
		p = append(p, &TCPStateCollector{})
	}
	if cfg.CollectListen {
		p = append(p, &ListenSocketCollector{})
	}
	if cfg.CollectTopTalkers {
		p = append(p, &TopTalkerCollector{})
	}

	return p
}
