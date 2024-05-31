package api_obj

import (
	"time"
)

type PodMetrics struct {
	Timestamp time.Time     `json:"timestamp" yaml:"timeStamp"`
	Window    time.Duration `json:"window" yaml:"window"`

	ContainerMetrics []ContainerMetrics `json:"containers" yaml:"containers"`
}

type ResourceList struct {
	CPUPercent    float64
	MemoryUsage   uint64
	MemoryPercent float64
}

type ContainerMetrics struct {
	Name  string       `json:"name" yaml:"name"`
	Usage ResourceList `json:"resourceList" yaml:"resourceList"`
}
