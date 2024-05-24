package api_obj

import (
	"minik8s/pkg/api_obj/obj_inner"
)

// 用于将function文件上传至服务器
type FunctionWrap struct {
	Func    Function
	Content []byte
}

type Function struct {
	Metadata  obj_inner.ObjectMeta `json:"metadata" yaml:"metadata"`
	FilePath  string               `json:"filePath" yaml:"filePath"`
	Coeff     string
}
