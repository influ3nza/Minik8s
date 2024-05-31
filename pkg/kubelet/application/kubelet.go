package application

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/config/kubelet"
	"minik8s/pkg/config/monitor"
	"minik8s/pkg/kubelet/util"
	"minik8s/pkg/message"
	"minik8s/pkg/network"
	"os"
	"os/signal"
	"strconv"
	"syscall"
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
	server.registerNodeToApiServerGetPods(node)
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

func (server *Kubelet) registerNodeToApiServerGetPods(node *api_obj.Node) {
	request, err := network.GetRequest(server.ApiServerAddress + apiserver.API_get_pods_by_node_force_prefix + node.NodeMetadata.Name)
	if err != nil {
		fmt.Println("Get All Pods to Initialize Node Failed, ", err.Error())
	}
	fmt.Println("Resp is ", request)
	list := gjson.Parse(request).Array()
	for _, p := range list {
		pod := &api_obj.Pod{}
		err = json.Unmarshal([]byte(p.String()), pod)
		if err != nil {
			fmt.Println("Unmarshal Error At GetPod To Initialize line 98 ", err.Error())
			continue
		}
		fmt.Println("Pod Name is ", pod.MetaData.Name, " Pod Ns is ", pod.MetaData.NameSpace)
		util.RegisterPod(pod.MetaData.Name, pod.MetaData.NameSpace)
	}
}

func (server *Kubelet) registerHandler() {
	server.Router.GET(kubelet.GetMetrics, server.GetPodMetrics)
	server.Router.DELETE(kubelet.DelPod, server.DelPod)
	server.Router.POST(kubelet.AddPod, server.AddPod)
	server.Router.GET(kubelet.GetCpuAndMem, server.GetNodeCPUAndMem)

	//PV
	server.Router.POST(kubelet.MountNfs, server.MountNfs)
	server.Router.DELETE(kubelet.UnmountNfs, server.UnmountNfs)
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

func (server *Kubelet) unregisterNodeToMonitor() {
	targetUrl := monitor.Server + monitor.UnRegisterNodePrefix
	hostname, _ := os.Hostname()
	targetUrl += hostname
	data, err := network.DelRequest(targetUrl)
	if err != nil {
		fmt.Println("unregisterNodeToMonitor Failed, ", err.Error())
		return
	}
	fmt.Println("unregisterNodeToMonitor Success, ", data)
}

func (server *Kubelet) Run() {
	server.register()
	server.registerHandler()
	go server.GetPodStatus()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	go func() {
		<-sigChan
		server.unregisterNodeToMonitor()
		os.Exit(0)
	}()

	err := server.Router.Run(fmt.Sprintf(":%d", server.Port))
	if err != nil {
		return
	}
}
