package api_obj

import (
	"minik8s/pkg/api_obj/obj_inner"
)

type ReplicaSet struct {
	ApiVersion string               `json:"apiVersion" yaml:"apiVersion"`
	Kind       string               `json:"kind" yaml:"kind"`
	MetaData   obj_inner.ObjectMeta `json:"metaData" yaml:"metadata"`
}

type ReplicaSetSpec struct {
	Replicas int
	Selector //TODO:
	Template //TODO:
}
