package api_obj

import (
	"minik8s/pkg/api_obj/obj_inner"

	"github.com/containerd/containerd"
)

type Pod struct {
	ApiVersion string               `json:"apiVersion" yaml:"apiVersion"`
	Kind       string               `json:"kind" yaml:"kind"`
	MetaData   obj_inner.ObjectMeta `json:"metaData" yaml:"metadata"`
	Spec       PodSpec              `json:"spec" yaml:"spec"`
	PodStatus  PodStatus
}

type PodSpec struct {
	Containers    []Container        `json:"containers" yaml:"containers"`
	Volumes       []obj_inner.Volume `json:"volumes" yaml:"volumes"`
	NodeName      string             `json:"nodeName" yaml:"nodeName"`
	NodeSelector  map[string]string  `json:"nodeSelector" yaml:"nodeSelector"`
	RestartPolicy string             `json:"restartPolicy" yaml:"restartPolicy"`
}

type Container struct {
	Name         string                         `json:"name" yaml:"name"`
	Image        obj_inner.Image                `json:"image" yaml:"image"`
	EntryPoint   obj_inner.EntryPoint           `json:"entryPoint" yaml:"entryPoint"`
	Ports        []obj_inner.ContainerPort      `json:"ports" yaml:"ports"`
	Env          []obj_inner.EnvVar             `json:"env" yaml:"env"`
	VolumeMounts []obj_inner.VolumeMount        `json:"volumeMounts" yaml:"volumeMounts"`
	Resources    obj_inner.ResourceRequirements `json:"resources" yaml:"resources"`
}

type PodStatus struct {
	PodIP          string              `json:"podIP" yaml:"podIP"`
	Phase          string              `json:"phase" yaml:"phase"`
	ContainerTypes []containerd.Status `json:"containerTypes" yaml:"containerTypes"`
}

type PodList struct {
}

//func (receiver *Pod) GetPodId() string {
//	return receiver.MetaData.UUID
//}
