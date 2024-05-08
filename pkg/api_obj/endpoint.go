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

	SrvIP   string `json:"srvIP" yaml:"srvIP"`
	SrvPort string `json:"srvPort" yaml:"srvPort"`

	PodUUID  string   `json:"PodUUID"`
	PodIP    string   `json:"PodIP"`
	PodPorts []string `json:"PodPort"`
}
