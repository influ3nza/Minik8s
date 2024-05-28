package container_manager

import (
	"github.com/containerd/containerd"
	"time"
)

type VolumeMap struct {
	Host_      string
	Container_ string
	Subdir_    string
	Type_      string
}

type MetricsCollection struct {
	begin      time.Time
	task       containerd.Task
	lastTime   time.Time
	lastCPU    uint64
	CPUPercent uint64
	memory     uint64
}

var IDLength int = 64
