package api_obj

import (
	"minik8s/pkg/api_obj/obj_inner"
)

type ScalingPolicyType string

const (
	PodsPolicy    ScalingPolicyType = "Pods"
	PercentPolicy ScalingPolicyType = "Percent"
)

type HPA struct {
	ApiVersion string               `json:"apiVersion" yaml:"apiVersion"`
	Kind       string               `json:"kind" yaml:"kind"`
	MetaData   obj_inner.ObjectMeta `json:"metaData" yaml:"metadata"`
	Spec       HPASpec              `json:"spec" yaml:"spec"`
	Status     HPAStatus            `json:"status" yaml:"status"`
}

type HPASpec struct {
	MinReplicas int                  `json:"minReplicas" yaml:"minReplicas"`
	MaxReplicas int                  `json:"maxReplicas" yaml:"maxReplicas"`
	Selector    map[string]string    `json:"selector" yaml:"selector"`
	Metrics     HPAMetrics           `json:"metrics" yaml:"metrics"`
	Policy      *ScalingPolicyType   `json:"policy" yaml:"policy"`
	Workload    obj_inner.ObjectMeta `json:"workload" yaml:"workload"`
}

type HPAMetrics struct {
	CPUPercent float64 `yaml:"cpuPercent" json:"cpuPercent"`
	MemPercent float64 `yaml:"memPercent" json:"memPercent"`
}

type HPAStatus struct {
	CurReplicas int     `yaml:"currentReplicas" json:"currentReplicas"`
	CurCpu      float64 `yaml:"curCpu" json:"curCpu"`
	CurMemory   float64 `yaml:"curMemory" json:"curMemory"`
}
