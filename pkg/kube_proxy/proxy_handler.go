package kube_proxy

import "minik8s/pkg/message"

func (m *ProxyManager) HandleMsg(msg *message.Message) {
	switch msg.Type {
	case message.SRV_CREATE:
	case message.SRV_DELETE:
	case message.EP_ADD:
	case message.EP_DELETE:
	}
}

func (m *ProxyManager) handleSrvCreate(content string) error {

	return nil
}
