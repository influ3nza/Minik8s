package app

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/etcd"
	"minik8s/pkg/message"
	"minik8s/tools"
)

type ApiServer struct {
	router   *gin.Engine
	EtcdWrap *etcd.EtcdWrap
	port     int32
	Producer *message.MsgProducer
}

// 在进行测试/实际运行时，第1步调用此函数。
func CreateApiServerInstance(c *config.ServerConfig) (*ApiServer, error) {
	router := gin.Default()
	router.SetTrustedProxies(c.TrustedProxy)

	wrap, err := etcd.CreateEtcdInstance(c.EtcdEndpoints, c.EtcdTimeout)
	if err != nil {
		fmt.Printf("create etcd instance failed, err:%v\n", err)
		return nil, err
	}

	producer := message.NewProducer()

	return &ApiServer{
		router:   router,
		EtcdWrap: wrap,
		port:     c.Port,
		Producer: producer,
	}, nil
}

// 建议放在其他的包内，使项目结构更整齐
func serverHelloWorld(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "hello world from apiserver!",
	})
}

// 将所有的接口在此函数内进行绑定
func (s *ApiServer) Bind() {
	s.router.GET("/hello", serverHelloWorld)

	s.router.GET(config.API_get_nodes, s.GetNodes)
	s.router.GET(config.API_get_node, s.GetNode)
	s.router.POST(config.API_add_node, s.AddNode)

	s.router.GET(config.API_get_pods, s.GetPods)
	s.router.POST(config.API_update_pod, s.UpdatePod)
	s.router.POST(config.API_add_pod, s.AddPod)

	s.router.POST(config.API_add_service, s.AddService)

	s.router.POST(config.API_add_endpoint, s.AddEndpoint)
	s.router.DELETE(config.API_delete_endpoint, s.DeleteEndpoint)
}

// 在进行测试/实际运行时，第2步调用此函数。默认端口为8080
func (s *ApiServer) Run() error {
	s.Bind()
	tools.Apiserver_boot_finished = true
	err := s.router.Run(fmt.Sprintf(":%d", s.port))
	return err
}
