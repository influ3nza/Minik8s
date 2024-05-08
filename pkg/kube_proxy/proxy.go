package kube_proxy

import (
	"minik8s/pkg/api_obj"
	"minik8s/pkg/message"

	"github.com/moby/ipvs"
)

type ProxyManager struct {
	Consumer    *message.MsgConsumer
	IpvsHandler *ipvs.Handle
	Services    *Services
}

type Manager interface {
	CreateService(srv *api_obj.Service)
	DeleteService(ip string, port int)
}

func (m *ProxyManager) CreateService(srv *api_obj.Service) {

}
