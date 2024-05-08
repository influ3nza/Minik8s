package kube_proxy

import (
	"github.com/moby/ipvs"
	"minik8s/pkg/message"
)

type ProxyManager struct {
	Consumer    *message.MsgConsumer
	IpvsHandler *ipvs.Handle
}

type Manager interface {
	CreateService()
}
