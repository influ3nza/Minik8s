package main

import (
	"fmt"
	"minik8s/pkg/kube_proxy"
)

func main() {
	ipv4, _ := kube_proxy.GetLocalIP()
	fmt.Printf("ipv4 is <%s>", ipv4)
	port, err := kube_proxy.GetFreePort()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(port)
}
