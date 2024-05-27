package app

import (
	"encoding/json"
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/message"
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

	if wf_name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/apiserver/AddWorkflow] Empty workflow name or namespace",
		})
		return
	}

	//查找是否已经存在相同的workflow
	e_key := apiserver.ETCD_workflow_prefix + wf_name
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

func (s *ApiServer) DeleteWorkflow(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	if name == "" || namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/apiserver/DeleteWorkflow] Empty name or namespace.",
		})
		return
	}

	e_key := apiserver.ETCD_workflow_prefix + namespace + "/" + name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/apiserver/DeleteWorkflow] Failed to get from etcd, " + err.Error(),
		})
		return
	}

	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/apiserver/DeleteWorkflow] Found zero or more than one workflow.",
		})
		return
	}

	err = s.EtcdWrap.Del(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/apiserver/DeleteWorkflow] Failed to delete from etcd, " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": "Delete workflow success",
	})
}

func (s *ApiServer) ExecWorkflow(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/apiserver/ExecWorkflow] Empty workflow name.",
		})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/apiserver/ExecWorkflow]" + err.Error(),
		})
		return
	}

	req_str, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/apiserver/ExecWorkflow]" + err.Error(),
		})
		return
	}

	e_key := apiserver.ETCD_workflow_prefix + name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/apiserver/ExecWorkflow] Failed to get from etcd, " + err.Error(),
		})
		return
	}

	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/apiserver/ExecWorkflow] Found zero or more than one wf.",
		})
		return
	}

	wf := &api_obj.Workflow{}
	err = json.Unmarshal([]byte(res[0].Value), wf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/apiserver/ExecWorkflow] Failed to unmarshal data, " + err.Error(),
		})
		return
	}

	wf.Spec.StartCoeff = string(req_str)

	wf_str, err := json.Marshal(wf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/apiserver/ExecWorkflow] Failed to marshal data, " + err.Error(),
		})
		return
	}

	s_msg := &message.Message{
		Type:    message.WF_EXEC,
		Content: string(wf_str),
	}
	s.Producer.Produce(message.TOPIC_Serverless, s_msg)

	c.JSON(http.StatusOK, gin.H{
		"data": "Sent exec request.",
	})
}

func (s *ApiServer) CheckWorkflow(c *gin.Context) {
	wf := &api_obj.Workflow{}
	err := c.ShouldBind(wf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/apiserver/CheckWorkflow] Failed to parse data, " + err.Error(),
		})
		return
	}

	fmt.Printf("[apiserver/CheckWorkflow] %v\n", wf)

	funcMap := make(map[string]api_obj.Function)
	for _, node := range wf.Spec.Nodes {
		if node.Type == api_obj.WF_Fork {
			continue
		}

		e_key := apiserver.ETCD_function_prefix + node.FuncSpec.Name
		res, err := s.EtcdWrap.Get(e_key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[ERR/apiserver/CheckWorkflow] Failed to get from etcd, " + err.Error(),
			})
			return
		}
		if len(res) != 1 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[ERR/apiserver/CheckWorkflow] Found zero or more than one function.",
			})
			return
		}

		f := &api_obj.Function{}
		err = json.Unmarshal([]byte(res[0].Value), f)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[ERR/apiserver/CheckWorkflow] Failed to unmarshal data, " + err.Error(),
			})
			return
		}

		funcMap[node.FuncSpec.Name] = *f
		fmt.Println(node.FuncSpec.Name)
	}

	map_str, err := json.Marshal(funcMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/apiserver/CheckWorkflow] Failed to marshal data, " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": string(map_str),
	})
}
