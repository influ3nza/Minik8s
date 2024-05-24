package monitor

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"minik8s/pkg/api_obj"
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

func (mc *MonitorController) registerHandler() {
	mc.Router.POST()
}

func Run() error {
	mc := InitController()
	if mc == nil {
		fmt.Println("Start Failed")
		return fmt.Errorf("gin engine init failed")
	}

}
