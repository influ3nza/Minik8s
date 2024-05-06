package application

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/kubelet/util"
	"minik8s/pkg/message"
	"minik8s/pkg/network"
	"os"
	"strconv"
	"time"
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
	hostName, _ := os.Hostname()
	node := api_obj.Node{
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
				ExternalIp: "",
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

	nodeJson, _ := json.Marshal(node)
	request, s, err := network.PostRequest(server.ApiServerAddress+"/nodes/add", nodeJson)
	if err != nil {
		fmt.Println("Send Register At line 54 ", err.Error())
		return
	}
	fmt.Println("Response data is ", request)
	fmt.Println("Response Err is ", s)
}

func (server *Kubelet) registerHandler() {
	server.Router.GET(util.GetMatrix, server.GetPodMatrix)
	server.Router.DELETE(util.DelPod, server.DelPod)
	server.Router.POST(util.AddPod, server.AddPod)
}

func InitKubeletDefault() *Kubelet {
	router := gin.Default()
	port, _ := strconv.Atoi(util.Port)
	producer := message.NewProducer()
	return &Kubelet{
		ApiServerAddress: util.ApiServer,
		Router:           router,
		IpAddress:        util.IpAddress,
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
