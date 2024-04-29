package api_obj

import (
	"time"
)

type PodMetrics struct {
	Timestamp time.Time     `json:"timestamp" yaml:"timeStamp"`
	Window    time.Duration `json:"window" yaml:"window"`

	// all containerMetrics are inn the same time window.
	Containers []ContainerMetrics `json:"containers" yaml:"containers"`
}

type ResourceList map[string]string

type ContainerMetrics struct {
	Name string `json:"name" yaml:"name"`
	// memory working set.
	Usage ResourceList `json:"resourceList" yaml:"resourceList"`
}
