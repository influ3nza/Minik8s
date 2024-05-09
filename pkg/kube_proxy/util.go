package kube_proxy

import (
	"net"
)

func GetLocalIP() (ipv4 string, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, addr := range addrs {
		ipNet, isIpNet := addr.(*net.IPNet)
		if isIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String()
				return
			}
		}
	}

	return
}
