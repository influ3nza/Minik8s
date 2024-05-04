package application

import (
	"github.com/gin-gonic/gin"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/kubelet/pod_manager"
	"minik8s/pkg/kubelet/util"
	"minik8s/pkg/message"
	"net/http"
	"sync"
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
		"ok": "[kubelet/AddPod] creating pod",
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		util.RegisterPod(pod.MetaData.Name, pod.MetaData.NameSpace)
		util.Lock(pod.MetaData.Name, pod.MetaData.NameSpace)
		err_ := pod_manager.AddPod(pod)
		util.UnLock(pod.MetaData.Name, pod.MetaData.NameSpace)
		if err_ != nil {
			msg := &message.Message{}
			server.Producer.Produce()
			util.UnRegisterPod(pod.MetaData.Name, pod.MetaData.NameSpace)
			wg.Done()
			return
		}
		wg.Done()
	}()
	wg.Wait()
	return
}

func (server *Kubelet) DelPod(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	util.Lock()
}
