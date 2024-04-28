package api_obj

import (
	"minik8s/pkg/api_obj/obj_inner"
	"time"
)

type PodMetrics struct {
	APIVersion string               `json:"apiVersion" yaml:"apiVersion"`
	MetaData   obj_inner.ObjectMeta `json:"metadata" yaml:"metadata"`

	Timestamp time.Time     `json:"timestamp" yaml:"timeStamp"`
	Window    time.Duration `json:"window" yaml:"window"`

	// all containerMetrics are inn the same time window.
	Containers []ContainerMetrics `json:"containers" yaml:"containers"`
}

type ResourceList map[string]uint64

type ContainerMetrics struct {
	Name string `json:"name" yaml:"name"`
	// memory working set.
	Usage ResourceList `json:"resourceList" yaml:"resourceList"`
}
