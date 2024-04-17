package app

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/etcd"
)

type ApiServer struct {
	router   *gin.Engine
	etcdWrap *etcd.EtcdWrap
	port     int32
}

func CreateApiServerInstance(c *config.ServerConfig) (*ApiServer, error) {
	engine := gin.Default()

	wrap, err := etcd.CreateEtcdInstance(c.EtcdEndpoints, c.EtcdTimeout)
	if err != nil {
		fmt.Printf("create etcd instance failed, err:%v\n", err)
		return nil, err
	}

	return &ApiServer{
		router:   engine,
		etcdWrap: wrap,
		port:     c.Port,
	}, nil
}

func serverHelloWorld(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "hello world from apiserver!",
	})
}

func (s *ApiServer) Bind() {
	s.router.GET("/hello", serverHelloWorld)
}

// 在进行测试/实际运行时，调用此函数。默认端口为8080
func (s *ApiServer) Run() error {
	s.Bind()
	err := s.router.Run(":%d", string(s.port))
	return err
}
