package application

import "github.com/gin-gonic/gin"

type Kubelet struct {
	ApiServerAddress string
	router           *gin.Engine
	port             int32
}
