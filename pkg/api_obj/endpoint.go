package api_obj

import (
	"minik8s/pkg/api_obj/obj_inner"
)

type Endpoint struct {
	MetaData obj_inner.ObjectMeta `json:"metaData" yaml:"metadata"`

	PodUUID string `json:"PodUUID"`
	PodIP   string `json:"PodIP"`
	PodPort int32  `json:"PodPort"`

	SrvPort int32 `json:"SrvPort"`
}
