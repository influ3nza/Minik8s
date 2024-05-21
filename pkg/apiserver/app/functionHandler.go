package app

import (
	"encoding/json"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/config/apiserver"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *ApiServer) AddFunction(c *gin.Context) {
	fw := &api_obj.FunctionWrap{}
	err := c.ShouldBind(fw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/AddFunction] Failed to parse function, " + err.Error(),
		})
		return
	}

	//将function存入etcd
	f := fw.Func
	if f.Metadata.Name == "" || f.FilePath == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddFunction] Empty function name or filepath.",
		})
		return
	}

	f_str, err := json.Marshal(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddFunction] Failed to marshal function, " + err.Error(),
		})
		return
	}

	e_key := apiserver.ETCD_function_prefix + f.Metadata.Name
	err = s.EtcdWrap.Put(e_key, f_str)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddFunction] Failed to write into etcd, " + err.Error(),
		})
		return
	}

	//TODO:将文件存在本地，并通知functionController。
}
