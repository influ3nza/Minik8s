package kube_proxy

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"minik8s/pkg/api_obj"
	kube_proxy "minik8s/pkg/config/kube_proxy"
)

func (m *ProxyManager) handleSrvCreate(c *gin.Context) {
	srv := &api_obj.Service{}
	err := c.ShouldBind(srv)
	if err != nil {
		fmt.Printf("[KubeProxy/Service Create] Failed to unmarshal service")
		return
	}

	err = m.CreateService(srv)
	if err != nil {
		fmt.Printf("[KubeProxy/Service Create] Failed to create service rules")
		return
	}
	return
}

func (m *ProxyManager) handleSrvDelete(c *gin.Context) {
	srv := &api_obj.Service{}
	err := c.ShouldBind(srv)
	if err != nil {
		fmt.Printf("[KubeProxy/Service Delete] Failed to unmarshal service")
		return
	}

	err = m.DelService(srv)
	if err != nil {
		fmt.Printf("[KubeProxy/Service Delete] Failed to del service rules")
		return
	}
	return
}

func (m *ProxyManager) handleEndpointAdd(c *gin.Context) {
	var list []api_obj.Endpoint
	err := c.ShouldBind(list)
	if err != nil {
		fmt.Printf("[KubeProxy/Endpoint Add] Failed to bind endpoint")
		return
	}

	for _, ep := range list {
		err = m.AddEndPoint(&ep)
		if err != nil {
			fmt.Printf("[KubeProxy/Endpoint Add] Failed to add endpoint rules")
			return
		}
	}

	return
}

func (m *ProxyManager) handleEndpointDelete(c *gin.Context) {
	var list []api_obj.Endpoint
	err := c.ShouldBind(list)
	if err != nil {
		fmt.Printf("[KubeProxy/Endpoint Delete] Failed to bind endpoint")
		return
	}

	for _, ep := range list {
		err = m.DelEndPoint(&ep)
		if err != nil {
			fmt.Printf("[KubeProxy/Endpoint Delete] Failed to delete endpoint rules")
			return
		}
	}

	return
}

func (m *ProxyManager) RegisterHandler() {
	m.Router.POST(kube_proxy.CreateService, m.handleSrvCreate)
	m.Router.POST(kube_proxy.DeleteService, m.handleSrvDelete)
	m.Router.POST(kube_proxy.AddEndpoint, m.handleEndpointAdd)
	m.Router.POST(kube_proxy.DeleteEndpoint, m.handleEndpointDelete)
}
