package application

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/kubelet/pod_manager"
	"minik8s/pkg/kubelet/util"
	"minik8s/pkg/message"
	"minik8s/pkg/network"
	"net/http"
	"time"
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
		if !util.Lock(pod.MetaData.Name, pod.MetaData.NameSpace) {
			errPod := api_obj.Pod{
				ApiVersion: "",
				Kind:       "",
				MetaData:   obj_inner.ObjectMeta{},
				Spec:       api_obj.PodSpec{},
				PodStatus: api_obj.PodStatus{
					PodIP: "error",
				},
			}
			errPodJson, _ := json.Marshal(errPod)
			msg := &message.Message{
				Type:    message.POD_CREATE,
				Content: string(errPodJson),
				Backup:  "This happened because of add not finished but deleted",
				Backup2: "",
			}
			server.Producer.Produce(message.TOPIC_ApiServer_FromNode, msg)
		}
		err_ := pod_manager.AddPod(pod)
		util.UnLock(pod.MetaData.Name, pod.MetaData.NameSpace)

		if err_ != nil {
			pod.PodStatus.PodIP = "error"
			msg := &message.Message{
				Type: message.POD_CREATE,
			}
			errPodJson, err_ := json.Marshal(*pod)
			if err_ != nil {
				errPod := api_obj.Pod{
					ApiVersion: "",
					Kind:       "",
					MetaData:   obj_inner.ObjectMeta{},
					Spec:       api_obj.PodSpec{},
					PodStatus: api_obj.PodStatus{
						PodIP: "error",
					},
				}
				errPodJson, _ = json.Marshal(errPod)
				msg.Content = string(errPodJson)
			} else {
				msg.Content = string(errPodJson)
			}
			server.Producer.Produce(message.TOPIC_ApiServer_FromNode, msg)
			for {
				if ok := util.UnRegisterPod(pod.MetaData.Name, pod.MetaData.NameSpace); ok == 0 || ok == 2 {
					break
				}
			}
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
		var ok int = 1
		for {
			if ok = util.UnRegisterPod(name, namespace); ok == 0 || ok == 2 {
				break
			}
		}
		msg := &message.Message{
			Type: message.POD_DELETE,
		}

		if ok == 2 {
			msg.Content = message.DEL_POD_NOT_EXIST
			server.Producer.Produce(message.TOPIC_ApiServer_FromNode, msg)
			return
		}

		err := pod_manager.DeletePod(name, namespace)
		if err != nil {
			msg.Content = message.DEL_POD_FAILED
			server.Producer.Produce(message.TOPIC_ApiServer_FromNode, msg)
			return
		}

		msg.Content = message.DEL_POD_SUCCESS
		server.Producer.Produce(message.TOPIC_ApiServer_FromNode, msg)
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

func (server *Kubelet) GetPodStatus() {
	for {
		time.Sleep(20 * time.Second)
		request, err := network.GetRequest(server.ApiServerAddress + "/pods/getAll")
		if err != nil {
			fmt.Println("Send Get RequestErr ", err.Error())
			return
		}

		list := gjson.Parse(request).Array()
		for _, p := range list {
			pod := &api_obj.Pod{}
			err = json.Unmarshal([]byte(p.String()), pod)
			if err != nil {
				fmt.Println("Unmarshal Error At GetPodStatus line 179 ", err.Error())
				continue
			}
			if !util.Lock(pod.MetaData.Name, pod.MetaData.NameSpace) {
				fmt.Println("Pod not Exist ", "Name is : ", pod.MetaData.Name, " Ns is : ", pod.MetaData.NameSpace)
				continue
			}
			res := pod_manager.MonitorPodContainers(pod.MetaData.Name, pod.MetaData.NameSpace)
			if res != obj_inner.Running {
				pod.PodStatus.Phase = res
				podJson, _ := json.Marshal(*pod)
				msg := &message.Message{
					Type:    message.POD_UPDATE,
					Content: string(podJson),
					Backup:  "",
					Backup2: "",
				}
				server.Producer.Produce(message.TOPIC_ApiServer_FromNode, msg)
			}
			util.UnLock(pod.MetaData.Name, pod.MetaData.NameSpace)
		}
	}
}
