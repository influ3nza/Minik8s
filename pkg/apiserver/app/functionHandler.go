package app

import (
	"encoding/json"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/kubectl/api"
	"minik8s/pkg/message"
	"minik8s/tools"
	"net/http"
	"os"

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
	f.Metadata.UUID = tools.NewUUID()
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
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddFunction] Failed to get from etcd, " + err.Error(),
		})
		return
	}
	if len(res) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/AddFunction] Function name already exists.",
		})
		return
	}

	err = s.EtcdWrap.Put(e_key, f_str)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddFunction] Failed to write into etcd, " + err.Error(),
		})
		return
	}

	//将文件存在本地，并通知functionController。
	path := "/mydata/" + f.Metadata.UUID + ".zip"
	file, err := os.Create(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddFunction] Failed to create file, " + err.Error(),
		})
		return
	}
	defer file.Close()

	// 2. Write the []byte data to the file
	_, err = file.Write(fw.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddFunction] Failed to write into file, " + err.Error(),
		})
		return
	}

	dirPath := "/mydata/" + f.Metadata.UUID
	err = api.DoUnzip(path, dirPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddFunction] Failed to unzip file, " + err.Error(),
		})
		return
	}

	//TODO:复制dockerfile

	//TODO:删除zip

	//向serverless组件发送消息
	s_msg := &message.Message{
		Type:    message.FUNC_CREATE,
		Content: string(f_str),
	}
	s.Producer.Produce(message.TOPIC_Serverless, s_msg)

	//返回200
	c.JSON(http.StatusOK, gin.H{
		"data": "[handler/AddFunction] Add function success",
	})
}

func (s *ApiServer) ExecFunction(c *gin.Context) {
	name := c.Param("name")
	coeff := c.Param("coeff")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/ExecFunction] Empty function name.",
		})
		return
	}

	e_key := apiserver.ETCD_function_prefix + name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/ExecFunction] Failed to get from etcd, " + err.Error(),
		})
		return
	}
	if len(res) != 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/ExecFunction] Found zero or more than one function.",
		})
		return
	}

	f := api_obj.Function{}
	err = json.Unmarshal([]byte(res[0].Value), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/ExecFunction] Failed to unmarshal data, " + err.Error(),
		})
		return
	}

	
}
