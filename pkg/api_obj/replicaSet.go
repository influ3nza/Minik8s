package api_obj

import (
	"minik8s/pkg/api_obj/obj_inner"
)

type ReplicaSet struct {
	ApiVersion string               `json:"apiVersion" yaml:"apiVersion"`
	Kind       string               `json:"kind" yaml:"kind"`
	MetaData   obj_inner.ObjectMeta `json:"metaData" yaml:"metadata"`
	Spec       ReplicaSetSpec       `json:"spec" yaml:"spec"`
	Status     ReplicaSetStatus     `json:"status" yaml:"status"`
}

type ReplicaSetSpec struct {
	Replicas int               `json:"replicas" yaml:"replicas"`
	Selector map[string]string `json:"selector" yaml:"selector"`
	Template PodTemplate       `json:"template" yaml:"template"`
}

type PodTemplate struct {
	Metadata obj_inner.ObjectMeta `json:"metadata" yaml:"metadata"`
	Spec     PodSpec              `json:"spec" yaml:"spec"`
}

type ReplicaSetStatus struct {
	Replicas      int    `json:"replicas" yaml:"replicas"`
	ReadyReplicas int    `json:"readyReplicas" yaml:"readyReplicas"`
	Status        string `json:"status" yaml:"status"`
	Message       string `json:"message" yaml:"message"`
}
