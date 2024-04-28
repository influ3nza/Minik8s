package api_obj

import (
	"minik8s/pkg/api_obj/obj_inner"
)

type Endpoint struct {
	//官方文档使用Cartesian Product进行ip集合和port集合的合并。
	//我们这里采用最原始的枚举方法。
	ApiVersion string               `json:"apiVersion" yaml:"apiVersion"`
	Kind       string               `json:"kind" yaml:"kind"`
	MetaData   obj_inner.ObjectMeta `json:"metadata" yaml:"metadata"`

	//where do you go?
	PodName string
	PodIP   string
	PodPort string

	//where do you come from?
	ServiceIP   string
	ServicePort string

	//if it hadn't been for Cotton Eye Joe,
	//i'd been married long time age,
	//where 'd you come from where 'd you go?
	//where 'd you come from Cotton Eye Joe?
}
