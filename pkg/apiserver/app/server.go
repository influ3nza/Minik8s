package app

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"minik8s/pkg/apiserver/controller/manager"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/dns"
	"minik8s/pkg/dns/dns_op"
	"minik8s/pkg/etcd"
	"minik8s/pkg/message"
	"minik8s/pkg/serverless/server"
	"minik8s/tools"
)

type ApiServer struct {
	router            *gin.Engine
	EtcdWrap          *etcd.EtcdWrap
	port              int32
	Producer          *message.MsgProducer
	Consumer          *message.MsgConsumer
	DnsService        *dns_op.DnsService
	ControllerManager manager.ControllerManager
	NodeIPMap         map[string]string
	S_server          *server.SL_server
}

// 在进行测试/实际运行时，第1步调用此函数。
func CreateApiServerInstance(c *apiserver.ServerConfig) (*ApiServer, error) {
	router := gin.Default()
	router.SetTrustedProxies(c.TrustedProxy)

	wrap, err := etcd.CreateEtcdInstance(c.EtcdEndpoints, c.EtcdTimeout)
	if err != nil {
		fmt.Printf("[ERR/Apiserver] create etcd instance failed, err:%v\n", err)
		return nil, err
	}

	producer := message.NewProducer()
	consumer, err := message.NewConsumer(message.TOPIC_ApiServer_FromNode, message.TOPIC_ApiServer_FromNode)
	if err != nil {
		fmt.Printf("[ERR/Apiserver] create kafka consumer instance failed, err:%v\n", err)
		return nil, err
	}

	dns_srv := dns.InitDnsService()
	ss, err := server.CreateNewSLServerInstance()
	if err != nil {
		fmt.Printf("[ERR/Apiserver] Failed to create serverless server.\n")
		return nil, err
	}

	cm, err := manager.CreateNewControllerManagerInstance()
	if err != nil {
		fmt.Printf("[ERR/Apiserver] Failed to create controller manager.\n")
		return nil, err
	}

	return &ApiServer{
		router:            router,
		EtcdWrap:          wrap,
		port:              c.Port,
		Producer:          producer,
		Consumer:          consumer,
		DnsService:        dns_srv,
		ControllerManager: cm,
		S_server:          ss,
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

	s.router.GET(apiserver.API_get_nodes, s.GetNodes)
	s.router.GET(apiserver.API_get_node, s.GetNode)
	s.router.POST(apiserver.API_add_node, s.AddNode)

	s.router.GET(apiserver.API_get_pods, s.GetPods)
	s.router.POST(apiserver.API_update_pod, s.UpdatePodScheduled)
	s.router.POST(apiserver.API_add_pod, s.AddPod)
	s.router.GET(apiserver.API_get_pods_by_node, s.GetPodsByNode)
	s.router.GET(apiserver.API_get_pods_by_node_force, )
	s.router.GET(apiserver.API_get_pod, s.GetPod)
	s.router.GET(apiserver.API_get_pods_by_namespace, s.GetPodsByNamespace)
	s.router.DELETE(apiserver.API_delete_pod, s.DeletePod)
	s.router.GET(apiserver.API_get_pod_metrics, s.GetPodMetrics)

	s.router.POST(apiserver.API_add_service, s.AddService)
	s.router.GET(apiserver.API_get_services, s.GetServices)
	s.router.GET(apiserver.API_get_service, s.GetService)
	s.router.DELETE(apiserver.API_delete_service, s.DeleteService)

	s.router.POST(apiserver.API_add_endpoint, s.AddEndpoint)
	s.router.DELETE(apiserver.API_delete_endpoints, s.DeleteEndpoints)
	s.router.DELETE(apiserver.API_delete_endpoint, s.DeleteEndpoint)
	s.router.GET(apiserver.API_get_endpoint, s.GetEndpoint)
	s.router.GET(apiserver.API_get_endpoint_by_service, s.GetEndpointsByService)

	s.router.GET(apiserver.API_get_replicasets, s.GetReplicaSets)        //need check, in replicasetHandler
	s.router.DELETE(apiserver.API_delete_replicaset, s.DeleteReplicaSet) //need check,in replicasetHandler
	s.router.POST(apiserver.API_update_replicaset, s.UpdateReplicaSet)
	s.router.POST(apiserver.API_add_replicaset, s.AddReplicaSet) //need check,in replicasetHandler

	s.router.GET(apiserver.API_scale_replicaset, s.ScaleReplicaSet)

	s.router.POST(apiserver.API_add_dns, s.AddDns)
	s.router.DELETE(apiserver.API_delete_dns, s.DeleteDns)
	s.router.GET(apiserver.API_get_dns)     //TODO
	s.router.GET(apiserver.API_get_all_dns) //TODO

	s.router.POST(apiserver.API_update_workflow) //TODO
	s.router.DELETE(apiserver.API_delete_workflow, s.DeleteWorkflow)
	s.router.POST(apiserver.API_exec_workflow, s.ExecWorkflow)
	s.router.POST(apiserver.API_add_workflow, s.AddWorkflow)
	s.router.GET(apiserver.API_get_workflow, s.GetWorkflow)
	s.router.POST(apiserver.API_check_workflow, s.CheckWorkflow)

	s.router.GET(apiserver.API_get_function) //TODO
	s.router.POST(apiserver.API_add_function, s.AddFunction)
	s.router.DELETE(apiserver.API_delete_function, s.DeleteFunction)
	s.router.GET(apiserver.API_exec_function, s.ExecFunction)
	s.router.GET(apiserver.API_find_function_ip, s.FindFunctionIp)
	s.router.GET(apiserver.API_get_function_res, s.GetFunctionRes)

	s.router.DELETE(apiserver.API_delete_registry, s.DeleteRegistry)
}

// 在进行测试/实际运行时，第2步调用此函数。默认端口为8080
func (s *ApiServer) Run() error {
	s.Bind()
	tools.NodesIpMap = make(map[string]string)
	tools.Apiserver_boot_finished = true

	//TODO:之后要在这里做容错。
	// s.EtcdWrap.DeleteByPrefix("/registry")
	tools.ClusterIpFlag = int32(s.ReadServiceMark())
	err := s.RefreshNodeIp()
	if err != nil {
		fmt.Printf("[ERR/apiserver] Failed tp refresh ip map.\n")
		return err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	go func() {
		<-sigChan
		s.Clean()
	}()

	go s.ControllerManager.Run()
	go s.Consumer.Consume([]string{message.TOPIC_ApiServer_FromNode}, s.MsgHandler)
	go s.S_server.Run()

	err = s.router.Run(fmt.Sprintf(":%d", s.port))
	return err
}

func (s *ApiServer) Clean() {
	fmt.Printf("[apiserver/CLEAN] Apiserver closing...\n")

	close(s.Consumer.Sig)
	close(s.Producer.Sig)
	s.Consumer.Consumer.Close()
	s.Producer.Producer.Close()
	os.Exit(0)
}
