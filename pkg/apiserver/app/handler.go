package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/apiserver/config"
)

func (s *ApiServer) GetNodes(c *gin.Context) {
	fmt.Printf("[apiserver/GetNodes] Try to get all nodes")

	res, err := s.EtcdWrap.GetByPrefix(config.ETCD_node_prefix)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[apiserver/GetNodes] Faled to get nodes, " + err.Error(),
		})
		return
	} else {
		var nodes []string
		for _, node := range res {
			nodes = append(nodes, node.Value)
		}

		c.JSON(http.StatusOK, gin.H{
			"data": nodes,
		})
	}
}

func (s *ApiServer) GetNode(c *gin.Context) {
	name := c.Param("name")
	fmt.Printf("[apiserver/GetNode] Try to get node: %s", name)

	if name == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "[apiserver/GetNode] Empty name string",
		})
		return
	} else {
		res, err := s.EtcdWrap.Get(config.ETCD_node_prefix + name)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "[apiserver/GetNode] Failed to get node, " + err.Error(),
			})
			return
		}

		if len(res) == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "[apiserver/GetNode] Failed to find node",
			})
			return
		}
		if len(res) != 1 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[apiserver/GetNode] Found more than one node",
			})
			return
		}

		result := res[0].Value
		c.JSON(http.StatusOK, gin.H{
			"data": result,
		})
		return
	}
}

func (s *ApiServer) AddNode(c *gin.Context) {
	fmt.Printf("[apiserver/AddNode] Try to add a node")

	var node api_obj.Node

	//TODO: post请求 要加JSON请求标头
	err := c.ShouldBind(&node)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[apiserver/AddNode] Failed to parse node, " + err.Error(),
		})

		return
	}

	//检查node各项参数
	//name是否重复
	node_name := node.GetName()
	res, err := s.EtcdWrap.Get(config.ETCD_node_prefix + node_name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/addNode] Get node failed, " + err.Error(),
		})
		return
	}

	if len(res) != 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/addNode] Node name already exists, " + err.Error(),
		})
		return
	}

	//初始化
	node.NodeMetadata.UUID = "default UUID"
	node.NodeStatus = api_obj.NodeStatus{}

	//parse， 此时的node已经是结构体了
	node_json, err := json.Marshal(node)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/addNode] Failed to marshal data, " + err.Error(),
		})
		return
	}

	err = s.EtcdWrap.Put(config.ETCD_node_prefix+node_name, node_json)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/addNode] Failed to write in etcd, " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": "[msgHandler/addNode] Add node success",
	})
}

func (s *ApiServer) AddPod(c *gin.Context) {
	//在etcd中创建一个新的pod对象，内容已从用户yaml文件中读取完毕。
	//TODO:
	new_pod := &api_obj.Pod{}
	err := c.ShouldBind(new_pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/addPod] Failed to parse pod from request, " + err.Error(),
		})
		return
	}

	new_pod_name := new_pod.MetaData.Name
	new_pod_namespace := new_pod.MetaData.NameSpace
	if new_pod_name == "" || new_pod_namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[msgHandler/addPod] Empty pod name or namespace",
		})
		return
	}

	//存入etcd
	//是否已经有同名pod
	res, err := s.EtcdWrap.Get(config.ETCD_pod_prefix + new_pod_name + "/" + new_pod_namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/addPod] Failed to get pod, " + err.Error(),
		})
		return
	}
	if len(res) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/addPod] Pod already exists",
		})
		return
	}

	//更新相关状态
	//TODO: 这里的UUID仍然为default。
	new_pod.MetaData.UUID = "default"
	new_pod_str, err := json.Marshal(new_pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/addPod] Failed to marshal pod, " + err.Error(),
		})
		return
	}

	//创建完成之后，通知scheduler分配node
	s.Producer.CallScheduleNode()
}

func (s *ApiServer) UpdatePod(c *gin.Context) {
	new_pod := &api_obj.Pod{}
	err := c.ShouldBind(new_pod)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/updatePod] Failed to parse pod, " + err.Error(),
		})
		return
	}

	//获取etcd中的原pod
	pod_name := new_pod.MetaData.Name
	pod_namespace := new_pod.MetaData.NameSpace
	fmt.Printf("[handler/updatePod] Try to get the original pod: %s/%s", pod_name, pod_namespace)

	if pod_name == "" || pod_namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/updatePod] Empty pod name or namespace",
		})
		return
	}

	e_key := config.ETCD_pod_prefix + pod_namespace + "/" + pod_name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/updatePod] Failed to get pod, " + err.Error(),
		})
		return
	}
	if len(res) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "[ERR/handler/updatePod] Failed to find pod",
		})
		return
	}
	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/updatePod] Found more than one pod",
		})
		return
	}

	old_pod := &api_obj.Pod{}
	err = json.Unmarshal([]byte(res[0].Value), old_pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/updatePod] Failed to unmarshal pod, " + err.Error(),
		})
		return
	}

	//TODO: 更新pod的状态，可能之后需要更新更多东西
	if new_pod.Spec.NodeName != "" {
		old_pod.Spec.NodeName = new_pod.Spec.NodeName
	}

	//存回etcd
	modified_pod, err := json.Marshal(old_pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/updatePod] Failed to marshall pod, " + err.Error(),
		})
		return
	}

	//使用原先的e_key，断言name和namespace不会发生变化
	err = s.EtcdWrap.Put(e_key, modified_pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/updatePod] Failed to write back etcd, " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": "[handler/updatePod] Update pod success",
	})
}
