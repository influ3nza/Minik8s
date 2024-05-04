package application

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/kubelet/pod_manager"
	"minik8s/pkg/kubelet/util"
	"minik8s/pkg/message"
	"net/http"
)

func (server *Kubelet) AddPod(c *gin.Context) {
	pod := &api_obj.Pod{}
	err := c.ShouldBind(pod)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[kubelet/AddPod] Failed to parse pod",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": "[kubelet/AddPod] creating pod",
	})

	go func() {
		util.RegisterPod(pod.MetaData.Name, pod.MetaData.NameSpace)
		util.Lock(pod.MetaData.Name, pod.MetaData.NameSpace)
		err_ := pod_manager.AddPod(pod)
		util.UnLock(pod.MetaData.Name, pod.MetaData.NameSpace)

		if err_ != nil {
			pod.PodStatus.PodIP = "error"
			msg := &message.Message{
				Type: message.POD_CREATE,
			}
			errPodJson, err_ := json.Marshal(*pod)
			if err_ != nil {
				err_pod := api_obj.Pod{
					ApiVersion: "",
					Kind:       "",
					MetaData:   obj_inner.ObjectMeta{},
					Spec:       api_obj.PodSpec{},
					PodStatus: api_obj.PodStatus{
						PodIP: "error",
					},
				}
				errPodJson, _ = json.Marshal(err_pod)
				msg.Content = string(errPodJson)
			} else {
				msg.Content = string(errPodJson)
			}
			server.Producer.Produce(message.TOPIC_ApiServer_FromNode, msg)
			util.UnRegisterPod(pod.MetaData.Name, pod.MetaData.NameSpace)
			return
		} else {
			msgPod, _ := json.Marshal(*pod)
			msg := &message.Message{
				Type:    message.POD_CREATE,
				Content: string(msgPod),
			}
			server.Producer.Produce(message.TOPIC_ApiServer_FromNode, msg)
			return
		}
	}()
	return
}

func (server *Kubelet) DelPod(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	c.JSON(http.StatusOK, gin.H{
		"data": "[kubectl/DelPod] deleting pod",
	})

	go func() {
		ok := util.Lock(name, namespace)
		msg := &message.Message{
			Type: message.POD_DELETE,
		}

		if !ok {
			msg.Content = message.DEL_POD_NOT_EXIST
			server.Producer.Produce(message.TOPIC_ApiServer_FromNode, msg)
			return
		}

		err := pod_manager.DeletePod(name, namespace)
		if err != nil {
			msg.Content = message.DEL_POD_FAILED
			server.Producer.Produce(message.TOPIC_ApiServer_FromNode, msg)
			util.UnLock(name, namespace)
			return
		}

		msg.Content = message.DEL_POD_SUCCESS
		server.Producer.Produce(message.TOPIC_ApiServer_FromNode, msg)
		util.UnLock(name, namespace)
		util.UnRegisterPod(name, namespace)
	}()

	return
}

func (server *Kubelet) GetPodMatrix(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	ok := util.Lock(name, namespace)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[kubectl/GetPodMatrix] Pod Not Exist",
		})
		return
	}

	res := pod_manager.GetPodMetrics(name, namespace)
	matrix, err := json.Marshal(*res)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[kubelet/GetPodMatrix] Marshal Error",
		})
		util.UnLock(name, namespace)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": string(matrix),
	})
	util.UnLock(name, namespace)
	return
}
