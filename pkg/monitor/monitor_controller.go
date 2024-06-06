package monitor

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/config/monitor"
	"net/http"
)

type MonitorController struct {
	Router *gin.Engine
}

func InitController() *MonitorController {
	eg := gin.Default()
	if eg == nil {
		return nil
	}
	return &MonitorController{
		Router: eg,
	}
}

func (mc *MonitorController) RegisterNode(c *gin.Context) {
	node := &api_obj.Node{}
	err := c.ShouldBind(node)
	if err != nil {
		fmt.Println("Get Node Req Failed, ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse node",
		})
		return
	}

	err = RegisterNode(node)
	if err != nil {
		fmt.Println("Register Node Failed, ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to register node",
		})
		return
	}
}

func (mc *MonitorController) RegisterPod(c *gin.Context) {
	pod := &api_obj.Pod{}
	err := c.ShouldBind(pod)
	if err != nil {
		fmt.Println("Get Pod Req Failed, ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse pod",
		})
		return
	}

	if pod.PodStatus.PodIP == "" || pod.MetaData.Labels == nil || pod.MetaData.Labels["metricsPort"] == "" {
		fmt.Println("Get Pod Ip or Metrics Port Failed")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to get ip or metricsPort",
		})
		return
	}

	err = RegisterPod(pod)
	if err != nil {
		fmt.Println("Register Pod Failed, ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to Register Pod",
		})
		return
	}
}

func (mc *MonitorController) UnRegisterNode(c *gin.Context) {
	hostname := c.Param("hostname")
	if hostname == "" {
		fmt.Println("No Hostname Here")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "no hostname in param",
		})
		return
	}

	err := UnRegisterNode(hostname)
	if err != nil {
		fmt.Println("UnRegister Node Failed, ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unregister node failed",
		})
		return
	}
}

func (mc *MonitorController) UnRegisterPod(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	if namespace == "" || name == "" {
		fmt.Println("No NameSpace or Name Here")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "no namespace or name in param",
		})
		return
	}

	err := UnRegisterPod(namespace, name)

	if err != nil {
		fmt.Println("UnRegister Pod Failed, ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unregister pod failed",
		})
		return
	}
}

func (mc *MonitorController) RegisterHandler() {
	mc.Router.POST(monitor.RegisterNode, mc.RegisterNode)
	mc.Router.POST(monitor.RegisterPod, mc.RegisterPod)
	mc.Router.DELETE(monitor.UnRegisterNode, mc.UnRegisterNode)
	mc.Router.DELETE(monitor.UnRegisterPod, mc.UnRegisterPod)
}
