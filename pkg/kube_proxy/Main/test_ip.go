package main

import (
	"fmt"
	"minik8s/pkg/kube_proxy"
)

func main() {
	ipv4, _ := kube_proxy.GetLocalIP()
	fmt.Printf("ipv4 is <%s>", ipv4)
}
