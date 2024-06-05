package api_obj

import "minik8s/pkg/api_obj/obj_inner"

type PVC_spec struct {
	AccessMode PV_access_mode    `json:"accessMode" yaml:"accessMode"`
	Selector   map[string]string `json:"selector" yaml:"selector"`
	BindPV     string            `json:"bindPV" yaml:"bindPV"`
}

type PVC struct {
	ApiVersion string               `json:"apiVersion" yaml:"apiVersion"`
	Kind       string               `json:"kind" yaml:"kind"`
	Metadata   obj_inner.ObjectMeta `json:"metadata" yaml:"metadata"`
	Spec       PVC_spec             `json:"spec" yaml:"spec"`
}
