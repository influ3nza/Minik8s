package api_obj

import "minik8s/pkg/api_obj/obj_inner"

type Dns struct {
	ApiVersion string               `json:"apiVersion" yaml:"apiVersion"`
	Kind       string               `json:"kind" yaml:"kind"`
	MetaData   obj_inner.ObjectMeta `json:"metaData" yaml:"metaData"`
	Host       string               `json:"host" yaml:"host"`
	Paths      []Path               `json:"paths" yaml:"paths"`
}

type Path struct {
	ServiceName string    `json:"serviceName" yaml:"serviceName"`
	Port        int32     `json:"port" yaml:"port"`
	SubPath     string    `json:"subPath" yaml:"subPath"`
	EndPoint    *Endpoint //todo apiServer 在加入dns的时候找到对应的ep 将指针放进去
}
