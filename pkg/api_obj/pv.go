package api_obj

import "minik8s/pkg/api_obj/obj_inner"

type PV_access_mode string

const (
	ReadWriteMany PV_access_mode = "ReadWriteMany"
	ReadWriteOnce PV_access_mode = "ReadWriteOnce"
)

type Nfs_bind struct {
	Path     string `json:"path" yaml:"path"`
	ServerIp string `json:"serverIp" yaml:"serverIp"`
}

type PV_spec struct {
	Nfs        Nfs_bind       `json:"nfs" yaml:"nfs"`
	AccessMode PV_access_mode `json:"accessMode" yaml:"accessMode"`
}

type PV struct {
	ApiVersion string               `json:"apiVersion" yaml:"apiVersion"`
	Kind       string               `json:"kind" yaml:"kind"`
	Metadata   obj_inner.ObjectMeta `json:"metadata" yaml:"metadata"`
	Spec       PV_spec              `json:"spec" yaml:"spec"`
}
