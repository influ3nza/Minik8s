package application

import (
	"github.com/gin-gonic/gin"
	"minik8s/pkg/message"
)

type Kubelet struct {
	ApiServerAddress string
	router           *gin.Engine
	port             int32
	Producer         *message.MsgProducer
}
