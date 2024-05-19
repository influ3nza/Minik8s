package app

import (
	"encoding/json"
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/config/apiserver"
	"minik8s/tools"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *ApiServer) AddWorkflow(c *gin.Context) {
	fmt.Printf("[apiserver/AddWorkflow] Try to add a workflow.\n")

	wf := &api_obj.Workflow{}
	err := c.ShouldBind(wf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/apiserver/AddWorkflow] Failed to parse workflow from request, " + err.Error(),
		})
		return
	}

	wf_name := wf.MetaData.Name
	wf_namespace := wf.MetaData.NameSpace

	if wf_name == "" || wf_namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/apiserver/AddWorkflow] Empty workflow name or namespace",
		})
		return
	}

	//查找是否已经存在相同的workflow
	e_key := apiserver.ETCD_workflow_prefix + wf_namespace + "/" + wf_name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/apiserver/AddWorkflow] Failed to get workflow, " + err.Error(),
		})
		return
	}

	if len(res) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/apiserver/AddWorkflow] Workflow already exists",
		})
		return
	}

	wf.MetaData.UUID = tools.NewUUID()

	//存入etcd
	wf_str, err := json.Marshal(wf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/apiserver/AddWorkflow] Failed to marshal workflow, " + err.Error(),
		})
		return
	}

	err = s.EtcdWrap.Put(e_key, wf_str)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/apiserver/AddWorkflow] Failed to write into etcd, " + err.Error(),
		})
		return
	}

	//返回200
	c.JSON(http.StatusCreated, gin.H{
		"data": "[apiserver/AddWorkflow] Create workflow success",
	})
}

func (s *ApiServer) GetWorkflow(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	if name == "" || namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/apiserver/GetWorkflow] Empty name or namespace.",
		})
		return
	}

	e_key := apiserver.ETCD_workflow_prefix + namespace + "/" + name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/apiserver/AddWorkflow] Failed to get from etcd, " + err.Error(),
		})
		return
	}

	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/apiserver/AddWorkflow] Found zero or more than one workflow.",
		})
		return
	}

	wf_str, err := json.Marshal(res[0].Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/apiserver/AddWorkflow] Failed to marshal data, " + err.Error(),
		})
		return
	}

	//返回200
	c.JSON(http.StatusCreated, gin.H{
		"data": wf_str,
	})
}
