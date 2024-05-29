package app

import (
	"encoding/json"
	"fmt"
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
	f.Metadata.NameSpace = apiserver.API_default_namespace
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

	//复制dockerfile等
	var p_path string
	dirPath += "/" + f.Metadata.Name + "/"
	if f.UseTemplate {
		p_path = tools.Func_template_path
	} else {
		p_path = f.FilePath
	}

	fmt.Printf("[AddFunction] P_path: %s, dirPath: %s\n", p_path, dirPath)

	api.DoCopy(p_path+"Dockerfile", dirPath+"Dockerfile")
	api.DoCopy(p_path+"requirements.txt", dirPath+"requirements.txt")
	api.DoCopy(p_path+"server.py", dirPath+"server.py")

	//删除zip
	err = os.Remove(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddFunction] Failed to delete file, " + err.Error(),
		})
		return
	}

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

	f := &api_obj.Function{}
	err = json.Unmarshal([]byte(res[0].Value), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/ExecFunction] Failed to unmarshal data, " + err.Error(),
		})
		return
	}

	if coeff != "nil" {
		f.Coeff = coeff
	}
	f_str, err := json.Marshal(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddFunction] Failed to marshal data, " + err.Error(),
		})
		return
	}

	//向serverless组件发送exec请求。
	m_msg := &message.Message{
		Type:    message.FUNC_EXEC,
		Content: string(f_str),
	}
	s.Producer.Produce(message.TOPIC_Serverless, m_msg)

	//返回200
	c.JSON(http.StatusOK, gin.H{
		"data": "[handler/AddFunction] Add function success",
	})
}

func (s *ApiServer) FindFunctionIp(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/FindFunctionIp] Empty function name.",
		})
		return
	}

	pack, err := s.GetPodsOfFunction(name)
	fmt.Printf("[handler/FindFunctionIp] Get ip: %s\n", pack)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/FindFunctionIp] Failed to get pod ip, " + err.Error(),
		})
		return
	}

	//检查是否有可用ip，如果没有，在这里进行扩容。
	if len(pack) == 0 {
		_, err = s.U_ScaleReplicaSet(name, 3)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[ERR/handler/FindFunctionIp] Failed to send scale rs request, " + err.Error(),
			})
			return
		}
	}

	pack_str := "["
	for id, ip := range pack {
		pack_str += "\"" + ip + "\""

		//返回值以逗号隔开
		if id < len(pack)-1 {
			pack_str += ","
		}
	}

	pack_str += "]"

	fmt.Printf("[handler/FindFunctionIp] Parse: %s\n", pack_str)

	//返回200
	c.JSON(http.StatusOK, gin.H{
		"data": pack_str,
	})
}

func (s *ApiServer) GetFunctionRes(c *gin.Context) {
	//TODO:通过文件路径获取结果。
}

func (s *ApiServer) DeleteFunction(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/DeleteFunction] Empty function name.",
		})
		return
	}

	e_key := apiserver.ETCD_function_prefix + name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/DeleteFunction] Failed to get from etcd, " + err.Error(),
		})
		return
	}

	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/DeleteFunction] Found zero or more than one function",
		})
		return
	}

	err = s.EtcdWrap.Del(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/DeleteFunction] Failed to delete from etcd, " + err.Error(),
		})
		return
	}

	f := &api_obj.Function{}
	err = json.Unmarshal([]byte(res[0].Value), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/DeleteFunction] Failed to unmarshal data, " + err.Error(),
		})
		return
	}

	dirPath := "/mydata/" + f.Metadata.UUID
	err = os.RemoveAll(dirPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/DeleteFunction] Failed to delete file, " + err.Error(),
		})
		return
	}

	//向serverless组件发送exec请求。
	m_msg := &message.Message{
		Type:    message.FUNC_DEL,
		Content: string(res[0].Value),
	}
	s.Producer.Produce(message.TOPIC_Serverless, m_msg)

	//返回200
	c.JSON(http.StatusOK, gin.H{
		"data": "Delete function success",
	})
}
