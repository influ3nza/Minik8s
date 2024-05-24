package api_obj

import "minik8s/pkg/api_obj/obj_inner"

type Job struct {
	Spec     JobSpec              `yaml:"jobspec" json:"jobspec"`
	MetaData obj_inner.ObjectMeta `yaml:"metaData" json:"metaData"`
	Status   JobStatus            `yaml:"status" json:"status"`
}

type JobSpec struct {
}

type JobStatus struct {
}
