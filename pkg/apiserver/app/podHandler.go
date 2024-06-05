package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/config/kubelet"
	"minik8s/pkg/message"
	"minik8s/pkg/network"
	"minik8s/tools"
)

func (s *ApiServer) GetPods(c *gin.Context) {
	fmt.Printf("[apiserver/GetPods] Try to get all pods.\n")

	res, err := s.EtcdWrap.GetByPrefix(apiserver.ETCD_pod_prefix)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[apiserver/GetPods] Failed to get pods, " + err.Error(),
		})
		return
	} else {
		var pods = "["
		for id, pod := range res {
			pods += pod.Value

			//返回值以逗号隔开
			if id < len(res)-1 {
				pods += ","
			}
		}

		pods += "]"

		c.JSON(http.StatusOK, gin.H{
			"data": pods,
		})
	}
}

func (s *ApiServer) AddPod(c *gin.Context) {
	fmt.Printf("[apiserver/AddPod] Try to add a pod.\n")

	//在etcd中创建一个新的pod对象，内容已从用户yaml文件中读取完毕。
	new_pod := &api_obj.Pod{}

	err := c.ShouldBind(new_pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[apiserver/AddPod] Failed to parse pod from request, " + err.Error(),
		})
		return
	}

	if new_pod.MetaData.Labels == nil {
		new_pod.MetaData.Labels = make(map[string]string)
	}
	if new_pod.MetaData.Annotations == nil {
		new_pod.MetaData.Annotations = make(map[string]string)
	}

	new_pod_name := new_pod.MetaData.Name
	new_pod_namespace := new_pod.MetaData.NameSpace

	if new_pod_name == "" || new_pod_namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[apiserver/AddPod] Empty pod name or namespace",
		})
		return
	}

	fmt.Printf("[apiserver/AddPod] Pod name: %s\n", new_pod_name)

	//存入etcd
	//是否已经有同名pod
	e_key := apiserver.ETCD_pod_prefix + new_pod_namespace + "/" + new_pod_name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddPod] Failed to get pod, " + err.Error(),
		})
		return
	}
	if len(res) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/AddPod] Pod already exists",
		})
		return
	}

	//更新相关状态
	new_pod.MetaData.UUID = tools.NewUUID()
	//WARN:（测试专用）
	if !tools.Test_enabled {
		new_pod.PodStatus.Phase = obj_inner.Pending
	}
	new_pod_str, err := json.Marshal(new_pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/msgHandler/AddPod] Failed to marshal pod, " + err.Error(),
		})
		return
	}

	err = s.EtcdWrap.Put(e_key, new_pod_str)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/msgHandler/AddPod] Failed to write into etcd, " + err.Error(),
		})
		return
	}

	//创建完成之后，通知scheduler分配node
	msg := &message.Message{
		Type:    message.SCHED,
		Content: string(new_pod_str),
	}

	s.Producer.Produce(message.TOPIC_Scheduler, msg)

	//成功返回
	c.JSON(http.StatusCreated, gin.H{
		"data": "[msgHandler/AddPod] Create pod success",
	})
}

func (s *ApiServer) UpdatePodScheduled(c *gin.Context) {
	new_pod := &api_obj.Pod{}
	err := c.ShouldBind(new_pod)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/UpdatePod] Failed to parse pod, " + err.Error(),
		})
		return
	}

	//获取etcd中的原pod
	pod_name := new_pod.MetaData.Name
	pod_namespace := new_pod.MetaData.NameSpace
	fmt.Printf("[handler/UpdatePod] Try to get the original pod: %s/%s\n", pod_namespace, pod_name)

	if pod_name == "" || pod_namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/UpdatePod] Empty pod name or namespace",
		})
		return
	}

	e_key := apiserver.ETCD_pod_prefix + pod_namespace + "/" + pod_name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/UpdatePod] Failed to get pod, " + err.Error(),
		})
		return
	}
	if len(res) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "[ERR/handler/UpdatePod] Failed to find pod",
		})
		return
	}
	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/UpdatePod] Found more than one pod",
		})
		return
	}

	old_pod := &api_obj.Pod{}
	err = json.Unmarshal([]byte(res[0].Value), old_pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/UpdatePod] Failed to unmarshal pod, " + err.Error(),
		})
		return
	}

	//更新pod的状态，可能之后需要更新更多东西
	if new_pod.Spec.NodeName != "" {
		old_pod.Spec.NodeName = new_pod.Spec.NodeName
	}

	//存回etcd
	modified_pod, err := json.Marshal(old_pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/UpdatePod] Failed to marshall pod, " + err.Error(),
		})
		return
	}

	//使用原先的e_key，断言name和namespace不会发生变化
	err = s.EtcdWrap.Put(e_key, modified_pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/UpdatePod] Failed to write back etcd, " + err.Error(),
		})
		return
	}

	//返回node的ip地址
	e_key = apiserver.ETCD_node_ip_prefix + new_pod.Spec.NodeName
	res, err = s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/UpdatePod] Failed to get from etcd, " + err.Error(),
		})
		return
	} else if len(res) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/UpdatePod] Specified node not available.",
		})
		return
	} else if len(res) > 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/UpdatePod] Found more than one node.",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": res[0].Value + strconv.Itoa(int(kubelet.Port)),
	})

	fmt.Printf("[handler/UpdatePod] Update pod success.\n")

	//仅供测试使用。
	// if tools.Test_enabled {
	// 	tools.Test_finished = true
	// }
}

func (s *ApiServer) GetPodsByNode(c *gin.Context) {
	//fmt.Printf("[apiserver/GetPodsByNode] Try to get all pods from a node.\n")
	nodename := c.Param("nodename")
	e_key := apiserver.ETCD_pod_prefix

	res, err := s.EtcdWrap.GetByPrefix(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/GetPodsByNode] Failed to get from etcd, " + err.Error(),
		})
		return
	}

	pack := []api_obj.Pod{}
	for _, kv := range res {
		pod := &api_obj.Pod{}
		err = json.Unmarshal([]byte(kv.Value), pod)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[ERR/handler/GetPodsByNode] Failed to unmarshal data, " + err.Error(),
			})
			return
		}

		if pod.Spec.NodeName == nodename && pod.PodStatus.Phase == obj_inner.Running {
			pack = append(pack, *pod)
		}
	}

	data := "["
	for id, p := range pack {
		p_str, err := json.Marshal(p)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[ERR/handler/GetPodsByNode] Failed to marshal data, " + err.Error(),
			})
			return
		}
		data += string(p_str)

		if id < len(res)-1 {
			data += ","
		}
	}

	data += "]"

	c.JSON(http.StatusCreated, gin.H{
		"data": data,
	})
}

func (s *ApiServer) GetPodsByNodeForce(c *gin.Context) {
	nodename := c.Param("nodename")
	e_key := apiserver.ETCD_pod_prefix

	res, err := s.EtcdWrap.GetByPrefix(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/GetPodsByNode] Failed to get from etcd, " + err.Error(),
		})
		return
	}

	pack := []api_obj.Pod{}
	for _, kv := range res {
		pod := &api_obj.Pod{}
		err = json.Unmarshal([]byte(kv.Value), pod)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[ERR/handler/GetPodsByNode] Failed to unmarshal data, " + err.Error(),
			})
			return
		}

		//和上一个函数的唯一区别
		if pod.Spec.NodeName == nodename {
			pack = append(pack, *pod)
		}
	}

	data := "["
	for id, p := range pack {
		p_str, err := json.Marshal(p)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[ERR/handler/GetPodsByNode] Failed to marshal data, " + err.Error(),
			})
			return
		}
		data += string(p_str)

		if id < len(res)-1 {
			data += ","
		}
	}

	data += "]"

	c.JSON(http.StatusCreated, gin.H{
		"data": data,
	})
}

func (s *ApiServer) DeletePod(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if name == "" || namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/DeletePod] Empty name or namespace.",
		})
		return
	}

	e_key := apiserver.ETCD_pod_prefix + namespace + "/" + name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/DeletePod] Failed to get from etcd, " + err.Error(),
		})
		return
	}
	if len(res) != 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/DeletePod] Found zero or more than one pod.",
		})
		return
	}

	old_pod := &api_obj.Pod{}
	err = json.Unmarshal([]byte(res[0].Value), old_pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/DeletePod] Failed to unmarshal data, " + err.Error(),
		})
		return
	}

	err = s.EtcdWrap.Del(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/DeletePod] Failed to delete from etcd, " + err.Error(),
		})
		return
	}

	//给kubelet发消息。
	nodeIp := tools.NodesIpMap[old_pod.Spec.NodeName]
	uri := nodeIp + strconv.Itoa(int(kubelet.Port)) +
		kubelet.DelPod_prefix + namespace + "/" + name + "/" + old_pod.MetaData.Annotations["pause"]
	fmt.Printf("Send pod delete to : %s.\n", uri)
	_, err = network.DelRequest(uri)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/DeletePod] Failed to send DEL request, " + err.Error(),
		})
		return
	}

	//给endpoint controller发消息。
	ep_msg := &message.Message{
		Type:    message.POD_DELETE,
		Content: res[0].Value,
	}
	s.Producer.Produce(message.TOPIC_EndpointController, ep_msg)

	c.JSON(http.StatusOK, gin.H{
		"data": "[handler/DeletePod] Delete pod success.",
	})
}

func (s *ApiServer) GetPod(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/GetPod] Empty namespace or name.",
		})
		return
	}

	e_key := apiserver.ETCD_pod_prefix + namespace + "/" + name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/GetPod] Failed to get from etcd, " + err.Error(),
		})
		return
	}

	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/GetPod] Found zero or more than one pod.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": res[0].Value,
	})
}

func (s *ApiServer) GetPodsByNamespace(c *gin.Context) {
	namespace := c.Param("namespace")

	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/GetPodsByNamespace] Empty namespace.",
		})
		return
	}

	e_key := apiserver.ETCD_pod_prefix + namespace
	res, err := s.EtcdWrap.GetByPrefix(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/GetPodsByNamespace] Failed to get from etcd, " + err.Error(),
		})
		return
	}

	data := "["
	for id, p := range res {
		p_str, err := json.Marshal(p)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[ERR/handler/GetPodsByNamespace] Failed to marshal data, " + err.Error(),
			})
			return
		}
		data += string(p_str)

		if id < len(res)-1 {
			data += ","
		}
	}

	data += "]"

	c.JSON(http.StatusCreated, gin.H{
		"data": data,
	})
}

func (s *ApiServer) GetPodMetrics(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if name == "" || namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/GetPodMetrics] Empty namespace or name.",
		})
		return
	}

	pod := &api_obj.Pod{}
	e_key := apiserver.ETCD_pod_prefix + namespace + "/" + name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/GetPodMetrics] Failed to get pod, " + err.Error(),
		})
		return
	}
	if len(res) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/GetPodMetrics] Pod does not exist.",
		})
		return
	}

	err = json.Unmarshal([]byte(res[0].Value), pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/GetPodMetrics] Failed to unmarshal data, " + err.Error(),
		})
		return
	}

	uri := tools.NodesIpMap[pod.Spec.NodeName] +
		strconv.Itoa(int(kubelet.Port)) + kubelet.GetMetrics_prefix + namespace + "/" + name

	pod_metrics := &api_obj.PodMetrics{}
	err = network.GetRequestAndParse(uri, pod_metrics)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/GetPodMetrics] Failed to send GET request, " + err.Error(),
		})
		return
	}

	pod_metrics_str, err := json.Marshal(pod_metrics)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/GetPodMetrics] Failed to marshal data, " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": string(pod_metrics_str),
	})
}
