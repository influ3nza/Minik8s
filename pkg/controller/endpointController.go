package controller

import (
	"encoding/json"
	"fmt"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/controller/utils"
	"minik8s/pkg/message"
	"minik8s/pkg/network"
)

//TODO:需要做的内容：service的创建与删除，pod的创建、删除和修改

type EndpointController struct {
	Consumer *message.MsgConsumer
}

func (ec *EndpointController) PrintHandlerWarning() {
	fmt.Printf("[WARN/EndpointController] Error in message handler, the system may not be working properly!\n")
}

func (ec *EndpointController) OnAddService(pack string) {
	//拿到所有的pod
	uri := config.API_server_prefix + config.API_get_pods
	dataStr, errStr, err := network.GetRequest(uri)
	if err != nil {
		fmt.Printf("[ERR/EndpointController/OnAddService] GET request failed, %v.\n", err)
		ec.PrintHandlerWarning()
		return
	} else if errStr != "" {
		fmt.Printf("[ERR/EndpointController/OnAddService] GET request failed, %s.\n", errStr)
		ec.PrintHandlerWarning()
		return
	}

	var allPods []api_obj.Pod
	if dataStr == "" {
		fmt.Printf("[ERR/EndpointController/OnAddService] Not any pod available.\n")
		ec.PrintHandlerWarning()
	} else {
		err = json.Unmarshal([]byte(dataStr), &allPods)
		if err != nil {
			fmt.Printf("[ERR/EndpointController/OnAddService] Failed to unmarshal dataS, %s.\n", err)
			ec.PrintHandlerWarning()
			return
		}
	}

	//从msg中读取service
	srv := &api_obj.Service{}
	err = json.Unmarshal([]byte(pack), srv)
	if err != nil {
		fmt.Printf("[ERR/EndpointController/OnAddService] Failed to unmarshal service, " + err.Error())
		ec.PrintHandlerWarning()
		return
	}

	//比对每一个pod
	for _, pod := range allPods {
		if pod.PodStatus.Phase == obj_inner.Running &&
			utils.CompareLabels(srv.Spec.Selector, pod.MetaData.Labels) &&
			pod.MetaData.NameSpace == srv.MetaData.NameSpace {
			//创建一个endpoint
			err = utils.CreateEndpoint(*srv, pod)
			if err != nil {
				fmt.Printf("[ERR/EndpointController/OnAddService] Failed to create endpoint, " + err.Error())
				ec.PrintHandlerWarning()
				return
			}
		}
	}
}

func (ec *EndpointController) OnDeleteService(pack string) {
	//需要删除所有的endpoints

	//从msg中读取service
	srv := &api_obj.Service{}
	err := json.Unmarshal([]byte(pack), srv)
	if err != nil {
		fmt.Printf("[ERR/EndpointController/OnDeleteService] Failed to unmarshal service, " + err.Error())
		ec.PrintHandlerWarning()
		return
	}

	uri := config.API_server_prefix + config.API_delete_endpoint + srv.MetaData.NameSpace + "/" + srv.MetaData.Name
	_, errStr, err := network.DelRequest(uri)
	if err != nil {
		fmt.Printf("[ERR/EndpointController/OnDeleteService] DEL request failed, %v.\n", err)
		ec.PrintHandlerWarning()
		return
	} else if errStr != "" {
		fmt.Printf("[ERR/EndpointController/OnDeleteService] DEL request failed, %s.\n", errStr)
		ec.PrintHandlerWarning()
		return
	}
}

func (ec *EndpointController) OnUpdatePod(pack string) {

}

func (ec *EndpointController) OnDeletePod(pack string) {

}

func (ec *EndpointController) MsgHandler(msg *message.Message) {
	fmt.Printf("[EndpointController/MsgHandler] Received message from apiserver!\n")

	switch msg.Type {
	case message.ENDPOINT_SRV_CREATE:
		ec.OnAddService(msg.Content)
	case message.ENDPOINT_SRV_DELETE:
		ec.OnDeleteService(msg.Content)
	case message.ENDPOINT_POD_UPDATE:
		ec.OnUpdatePod(msg.Content)
	case message.ENDPOINT_POD_DELETE:
		ec.OnDeletePod(msg.Content)
	}

}

func (ec *EndpointController) Run() {
	go ec.Consumer.Consume([]string{message.TOPIC_EndpointController}, ec.MsgHandler)
}

func CreateEndpointControllerInstance() (*EndpointController, error) {
	consumer, err := message.NewConsumer(message.TOPIC_EndpointController, "default")
	return &EndpointController{
		Consumer: consumer,
	}, err
}
