package main

import (
	"fmt"
	"minik8s/pkg/controller"
	"minik8s/tools"
	"time"
)

func main() {
	tools.NodesIpMap = make(map[string]string)
	//TODO:仅供测试使用，需要取消注释。
	nodeaddr := "http://127.0.0.1:"
	// nodeaddr := node.GetInternelIp() + ":" + strconv.Itoa(int(node.NodeStatus.Addresses.Port))
	tools.NodesIpMap["node-example1"] = nodeaddr

	ep, err := controller.CreateEndpointControllerInstance()
	if err != nil {
		fmt.Printf("[Controller/MAIN] Failed to create ep controller.")
	}

	go ep.Run()

	for {
		time.Sleep(100 * time.Millisecond)
	}
}
