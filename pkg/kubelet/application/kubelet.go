package application

import (
	"encoding/json"
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/config/kubelet"
	"minik8s/pkg/config/monitor"
	"minik8s/pkg/kubelet/util"
	"minik8s/pkg/message"
	"minik8s/pkg/network"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Kubelet struct {
	ApiServerAddress string
	Router           *gin.Engine
	IpAddress        string
	Port             int32
	Producer         *message.MsgProducer
	TotalCpu         string
	TotalMem         string
	Label            map[string]string
}

func (server *Kubelet) register() {
	// fmt.Println(server.IpAddress)
	hostName, _ := os.Hostname()
	node := &api_obj.Node{
		APIVersion: "v1",
		NodeMetadata: obj_inner.ObjectMeta{
			Name: hostName,
			Labels: map[string]string{
				"name": hostName,
			},
		},
		Kind: "node",
		NodeStatus: api_obj.NodeStatus{
			Addresses: api_obj.Address{
				HostName:   hostName,
				ExternalIp: "10.119.13.178",
				InternalIp: server.IpAddress,
			},
			Condition: api_obj.Ready,
			Capacity: map[string]string{
				"cpu":    server.TotalCpu,
				"memory": server.TotalMem,
			},
			Allocatable: map[string]string{},
			UpdateTime:  time.Now(),
		},
	}

	server.registerNodeToApiServer(node)
	server.registerNodeToMonitor(node)
}

func (server *Kubelet) registerNodeToApiServer(node *api_obj.Node) {
	nodeJson, _ := json.Marshal(*node)
	request, err := network.PostRequest(server.ApiServerAddress+apiserver.API_add_node, nodeJson)
	if err != nil {
		fmt.Println("Send Register At line 62 ", err.Error())
		return
	}
	fmt.Println("Response data on register to apiServer is ", request)
}

func (server *Kubelet) registerNodeToMonitor(node *api_obj.Node) {
	nodeJson, _ := json.Marshal(*node)
	request, err := network.PostRequest(monitor.Server+monitor.RegisterNode, nodeJson)
	if err != nil {
		fmt.Println("Send Register  At line 69 ", err.Error())
	}
	fmt.Println("Response data on register to monitor is ", request)
}

func (server *Kubelet) registerHandler() {
	server.Router.GET(kubelet.GetMatrix, server.GetPodMatrix)
	server.Router.DELETE(kubelet.DelPod, server.DelPod)
	server.Router.POST(kubelet.AddPod, server.AddPod)

	//PV
	server.Router.POST(kubelet.MountNfs, server.MountNfs)
}

func InitKubeletDefault() *Kubelet {
	router := gin.Default()
	port, _ := strconv.Atoi(util.Port)
	producer := message.NewProducer()
	return &Kubelet{
		ApiServerAddress: util.ApiServer,
		Router:           router,
		IpAddress:        util.IpAddressMas,
		Port:             int32(port),
		Producer:         producer,
		TotalCpu:         util.Cpu,
		TotalMem:         util.Memory,
		Label:            map[string]string{},
	}
}

func InitKubelet(config util.KubeConfig) *Kubelet {
	router := gin.Default()
	producer := message.NewProducer()
	return &Kubelet{
		ApiServerAddress: config.ApiServer,
		Router:           router,
		IpAddress:        config.Ip,
		Port:             config.Port,
		Producer:         producer,
		TotalCpu:         config.TotalCpu,
		TotalMem:         config.TotalMem,
		Label:            config.Label,
	}
}

func (server *Kubelet) Run() {
	server.register()
	server.registerHandler()
	go server.GetPodStatus()
	err := server.Router.Run(fmt.Sprintf(":%d", server.Port))
	if err != nil {
		return
	}
}
