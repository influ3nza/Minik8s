package api_obj

import (
	"minik8s/pkg/api_obj/obj_inner"
)

type Endpoint struct {
	//官方文档使用Cartesian Product进行ip集合和port集合的合并。
	//我们这里采用最原始的枚举方法。
	ApiVersion string               `json:"apiVersion" yaml:"apiVersion"`
	Kind       string               `json:"kind" yaml:"kind"`
	MetaData   obj_inner.ObjectMeta `json:"metaData" yaml:"metadata"`

	SrvUUID string `json:"srvUUID" yaml:"srvUUID"`
	SrvIP   string `json:"srvIP" yaml:"srvIP"`
	SrvPort int32  `json:"srvPort" yaml:"srvPort"`

	PodUUID string `json:"PodUUID"`
	PodIP   string `json:"PodIP"`
	PodPort string `json:"PodPort"`
	Weight  int    `json:"Weight"`
}
