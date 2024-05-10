package kube_proxy

import (
	"encoding/json"
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/message"
)

func (m *ProxyManager) HandleMsg(msg *message.Message) {
	switch msg.Type {
	case message.SRV_CREATE:
		{
			err := m.handleSrvCreate(msg.Content)
			if err != nil {

			} else {

			}
		}
	case message.SRV_DELETE:
		{
			err := m.handleSrvDelete(msg.Content)
			if err != nil {

			} else {

			}
		}
	case message.EP_ADD:
		{
			err := m.handleEndpointAdd(msg.Content)
			if err != nil {

			} else {

			}
		}
	case message.EP_DELETE:
	}
}

func (m *ProxyManager) handleSrvCreate(content string) error {
	srv := &api_obj.Service{}
	err := json.Unmarshal([]byte(content), srv)
	if err != nil {
		return fmt.Errorf("[KubeProxy/Service Create] Failed to unmarshal service")
	}

	err = m.CreateService(srv)
	if err != nil {
		return fmt.Errorf("[KubeProxy/Service Create] Failed to create service rules")
	}
	return nil
}

func (m *ProxyManager) handleSrvDelete(content string) error {
	srv := &api_obj.Service{}
	err := json.Unmarshal([]byte(content), srv)
	if err != nil {
		return fmt.Errorf("[KubeProxy/Service Delete] Failed to unmarshal service")
	}

	err = m.DelService(srv)
	if err != nil {
		return fmt.Errorf("[KubeProxy/Service Delete] Failed to del service rules")
	}
	return nil
}

func (m *ProxyManager) handleEndpointAdd(content string) error {
	ep := &api_obj.Endpoint{}
	err := json.Unmarshal([]byte(content), ep)
	if err != nil {
		return fmt.Errorf("[KubeProxy/Endpoint Add] Failed to unmarshal endpoint")
	}

	err = m.AddEndPoint(ep)
	if err != nil {
		return fmt.Errorf("[KubeProxy/Endpoint Add] Failed to add endpoint rules")
	}
	return nil
}

func (m *ProxyManager) handleEndpointDelete(content string) error {
	ep := &api_obj.Endpoint{}
	err := json.Unmarshal([]byte(content), ep)
	if err != nil {
		return fmt.Errorf("[KubeProxy/Endpoint Delete] Failed to unmarshal endpoint")
	}

	err = m.DelEndPoint(ep)
	if err != nil {
		return fmt.Errorf("[KubeProxy/Endpoint Delete] Failed to delete endpoint rules")
	}
	return nil
}
