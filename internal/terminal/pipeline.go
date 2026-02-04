package terminal

import "vsysmon/internal/config"

func BuildSections(cfg config.Config) []MetricRenderer {
	var s []MetricRenderer

	if cfg.CollectLoad {
		s = append(s, &LoadRenderer{})
	}
	if cfg.CollectCPU {
		s = append(s, &RenderCPU{})
	}
	if cfg.CollectDisk {
		s = append(s, &RenderDisk{})
	}
	if cfg.CollectFS {
		s = append(s, &RenderFS{})
	}
	if cfg.CollectTCPStates {
		s = append(s, &RenderTCP{})
	}
	if cfg.CollectTopTalkers {
		s = append(s, &RenderProtoTop{}, &RenderFlowTop{})
	}
	if cfg.CollectListen {
		s = append(s, &RenderListen{})
	}
	return s
}
