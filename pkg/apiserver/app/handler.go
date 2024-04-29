package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/apiserver/config"
	"minik8s/tools"
)

func (s *ApiServer) GetNodes(c *gin.Context) {
	fmt.Printf("[apiserver/GetNodes] Try to get all nodes.\n")

	res, err := s.EtcdWrap.GetByPrefix(config.ETCD_node_prefix)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[apiserver/GetNodes] Failed to get nodes, " + err.Error(),
		})
		return
	} else {
		var nodes []string
		for id, node := range res {
			nodes = append(nodes, node.Value)

			//返回值以逗号隔开
			if id < len(res)-1 {
				nodes = append(nodes, ",")
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"data": nodes,
		})
	}
}

func (s *ApiServer) GetNode(c *gin.Context) {
	name := c.Param("name")
	fmt.Printf("[apiserver/GetNode] Try to get node: %s.\n", name)

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
	fmt.Printf("[apiserver/addNode] Try to add a node.\n")

	var node api_obj.Node

	err := c.ShouldBind(&node)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/apiserver/addNode] Failed to parse node, " + err.Error(),
		})

		return
	}

	//检查node各项参数
	//name是否重复，是否为空
	node_name := node.GetName()
	node_namespace := node.NodeMetadata.NameSpace

	if node_name == "" || node_namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[msgHandler/addNode] Empty node name or namespace.",
		})
		return
	}

	fmt.Printf("[apiserver/addNode] Node name: %s\n", node_name)

	res, err := s.EtcdWrap.Get(config.ETCD_node_prefix + node_name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/addNode] Get node failed, " + err.Error(),
		})
		return
	}

	if len(res) != 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/addNode] Node name already exists.",
		})
		return
	}

	//初始化
	node.NodeMetadata.UUID = "default UUID"
	//TODO: 这里是便于测试，之后需要重新书写
	node.NodeStatus = api_obj.NodeStatus{
		Condition: api_obj.Ready,
	}

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

func (s *ApiServer) GetPods(c *gin.Context) {
	fmt.Printf("[apiserver/GetPods] Try to add all pods.\n")

	res, err := s.EtcdWrap.GetByPrefix(config.ETCD_pod_prefix)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[apiserver/GetPods] Failed to get pods, " + err.Error(),
		})
		return
	} else {
		var pods []string
		for id, pod := range res {
			pods = append(pods, pod.Value)

			//返回值以逗号隔开
			if id < len(res)-1 {
				pods = append(pods, ",")
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"data": pods,
		})
	}
}

func (s *ApiServer) AddPod(c *gin.Context) {
	fmt.Printf("[apiserver/addPod] Try to add a pod.\n")

	//在etcd中创建一个新的pod对象，内容已从用户yaml文件中读取完毕。
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

	fmt.Printf("[apiserver/addPod] Pod name: %s\n", new_pod_name)

	//存入etcd
	//是否已经有同名pod
	e_key := config.ETCD_pod_prefix + new_pod_namespace + "/" + new_pod_name
	res, err := s.EtcdWrap.Get(e_key)
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
	new_pod.PodStatus.Phase = obj_inner.Pending
	new_pod_str, err := json.Marshal(new_pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/msgHandler/addPod] Failed to marshal pod, " + err.Error(),
		})
		return
	}

	err = s.EtcdWrap.Put(e_key, new_pod_str)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/msgHandler/addPod] Failed to write into etcd, " + err.Error(),
		})
		return
	}

	//创建完成之后，通知scheduler分配node
	s.Producer.CallScheduleNode(string(new_pod_str))

	//成功返回
	c.JSON(http.StatusCreated, gin.H{
		"data": "[msgHandler/addPod] Create pod success",
	})
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
	fmt.Printf("[handler/updatePod] Try to get the original pod: %s/%s\n", pod_namespace, pod_name)

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

	//测试的终点，到达这里就可以下班了
	fmt.Printf("[handler/updatePod] Update pod success.\n")

	//仅供测试使用。
	if tools.Test_enabled {
		tools.Test_finished = true
	}
}

func (s *ApiServer) AddService(c *gin.Context) {
	fmt.Printf("[apiserver/AddService] Try to add a service.\n")

	new_service := &api_obj.Service{}
	err := c.ShouldBind(new_service)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/AddService] Failed to parse service, " + err.Error(),
		})
		return
	}

	service_name := new_service.MetaData.Name
	service_namespace := new_service.MetaData.NameSpace

	if service_name == "" || service_namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/AddService] Empty service name or namespace",
		})
		return
	}

	fmt.Printf("[apiserver/AddService] Service name: %s\n", service_name)

	e_key := config.ETCD_service_prefix + service_namespace + "/" + service_name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddService] Get service failed, " + err.Error(),
		})
		return
	}

	if len(res) != 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddService] Service name already exists.",
		})
		return
	}

	//TODO:分配IP（检查IP）
	//TODO:分配UUID
	new_service.MetaData.UUID = "default_service"
	new_service.Status = api_obj.ServiceStatus{
		Condition: api_obj.SERVICE_PENDING,
	}

	//存入etcd
	service_str, err := json.Marshal(new_service)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddService] Failed to marshal service, " + err.Error(),
		})
		return
	}

	err = s.EtcdWrap.Put(e_key, service_str)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddService] Failed to write into etcd, " + err.Error(),
		})
		return
	}

	//TODO:创建endpoints
	//TODO:返回200
	//TODO:通知proxy，service已经创建
}
