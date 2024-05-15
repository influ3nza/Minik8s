package controller

import (
	"encoding/json"
	"fmt"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/controller/utils"
	"minik8s/pkg/message"
	"minik8s/pkg/network"
)

//需要做的内容：service的创建与删除，pod的创建、删除和修改

type EndpointController struct {
	Consumer *message.MsgConsumer
}

func (ec *EndpointController) PrintHandlerWarning() {
	fmt.Printf("[WARN/EndpointController] Error in message handler, the system may not be working properly!\n")
}

func (ec *EndpointController) OnAddService(pack string) {
	//拿到所有的pod
	uri := apiserver.API_server_prefix + apiserver.API_get_pods

	var allPods []api_obj.Pod
	err := network.GetRequestAndParse(uri, &allPods)
	if err != nil {
		fmt.Printf("[ERR/EndpointController/OnAddService] GET request failed, %v.\n", err)
		ec.PrintHandlerWarning()
		return
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
	ep_pod_list := []api_obj.Pod{}
	for _, pod := range allPods {
		if pod.PodStatus.Phase == obj_inner.Running &&
			utils.CompareLabels(srv.Spec.Selector, pod.MetaData.Labels) &&
			pod.MetaData.NameSpace == srv.MetaData.NameSpace {
			ep_pod_list = append(ep_pod_list, pod)
		}
	}

	//创建endpoints
	err = utils.CreateEndpoints([]api_obj.Service{*srv}, ep_pod_list)
	if err != nil {
		fmt.Printf("[ERR/EndpointController/OnAddService] Failed to create endpoint, " + err.Error())
		ec.PrintHandlerWarning()
		return
	}

	//仅供测试使用
	// if tools.Test_enabled {
	// 	tools.Test_finished = true
	// }
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

	err = utils.DeleteEndpoints(true, srv.MetaData.NameSpace+"/"+srv.MetaData.Name)
	if err != nil {
		fmt.Printf("[ERR/EndpointController/OnDeleteService] Failed to delete endpoint, " + err.Error())
		ec.PrintHandlerWarning()
		return
	}
}

func (ec *EndpointController) OnCreatePod(pack string) {
	//从msg中读取pod
	pod := &api_obj.Pod{}
	err := json.Unmarshal([]byte(pack), pod)
	if err != nil {
		fmt.Printf("[ERR/EndpointController/OnCreatePod] Failed to unmarshal pod, " + err.Error())
		ec.PrintHandlerWarning()
		return
	}

	//拿到所有service
	uri := apiserver.API_server_prefix + apiserver.API_get_services
	dataStr, err := network.GetRequest(uri)
	if err != nil {
		fmt.Printf("[ERR/EndpointController/OnCreatePod] GET request failed, %v.\n", err)
		ec.PrintHandlerWarning()
		return
	}

	var allSrvs []api_obj.Service
	if dataStr == "" {
		fmt.Printf("[ERR/EndpointController/OnCreatePod] Not any service available.\n")
		//WARN: 仅供测试使用。
		// if tools.Test_enabled {
		// 	tools.Test_finished = true
		// }
		return
	} else {
		err = json.Unmarshal([]byte(dataStr), &allSrvs)
		if err != nil {
			fmt.Printf("[ERR/EndpointController/OnCreatePod] Failed to unmarshal data, %s.\n", err)
			ec.PrintHandlerWarning()
			return
		}
	}

	//遍历service，寻找所在的所有service并且添加
	ep_srv_list := []api_obj.Service{}
	for _, srv := range allSrvs {
		if utils.CompareLabels(srv.MetaData.Labels, pod.MetaData.Labels) &&
			pod.MetaData.NameSpace == srv.MetaData.NameSpace {
			ep_srv_list = append(ep_srv_list, srv)
		}
	}

	//创建endpoints
	err = utils.CreateEndpoints(ep_srv_list, []api_obj.Pod{*pod})
	if err != nil {
		fmt.Printf("[ERR/EndpointController/OnCreatePod] Failed to create endpoint, " + err.Error())
		ec.PrintHandlerWarning()
		return
	}
}

func (ec *EndpointController) OnUpdatePod(pack string) {
	//此处必定发生ip地址修改。相当于先删除pod对应的所有endpoints再添加。
	//pack的内容为新的pod结构体。
	ec.OnDeletePod(pack)
	ec.OnCreatePod(pack)
}

func (ec *EndpointController) OnDeletePod(pack string) {
	//从msg中读取pod
	pod := &api_obj.Pod{}
	err := json.Unmarshal([]byte(pack), pod)
	if err != nil {
		fmt.Printf("[ERR/EndpointController/OnDeletePod] Failed to unmarshal pod, " + err.Error())
		ec.PrintHandlerWarning()
		return
	}

	//拿到所有service
	uri := apiserver.API_server_prefix + apiserver.API_get_services
	dataStr, err := network.GetRequest(uri)
	if err != nil {
		fmt.Printf("[ERR/EndpointController/OnDeletePod] GET request failed, %v.\n", err)
		ec.PrintHandlerWarning()
		return
	}

	var allSrvs []api_obj.Service
	if dataStr == "" {
		fmt.Printf("[ERR/EndpointController/OnDeletePod] Not any service available.\n")
		//WARN: 仅供测试使用。
		// if tools.Test_enabled {
		// 	tools.Test_finished = true
		// }
		return
	} else {
		err = json.Unmarshal([]byte(dataStr), &allSrvs)
		if err != nil {
			fmt.Printf("[ERR/EndpointController/OnDeletePod] Failed to unmarshal data, %s.\n", err)
			ec.PrintHandlerWarning()
			return
		}
	}

	//遍历service，寻找所在的所有service并且删除
	for _, srv := range allSrvs {
		if utils.CompareLabels(srv.MetaData.Labels, pod.MetaData.Labels) {
			//删除对应的endpoint
			suffix := srv.MetaData.NameSpace + "/" +
				srv.MetaData.Name + "-" + pod.MetaData.Name
			err = utils.DeleteEndpoints(false, suffix)
			if err != nil {
				fmt.Printf("[ERR/EndpointController/OnDeletePod] Failed to delete endpoint, " + err.Error())
				ec.PrintHandlerWarning()
				return
			}
		}
	}
}

func (ec *EndpointController) MsgHandler(msg *message.Message) {
	fmt.Printf("[EndpointController/MsgHandler] Received message from apiserver!\n")

	switch msg.Type {
	case message.SRV_CREATE:
		ec.OnAddService(msg.Content)
	case message.SRV_DELETE:
		ec.OnDeleteService(msg.Content)
	case message.POD_CREATE:
		ec.OnCreatePod(msg.Content)
	case message.POD_UPDATE:
		ec.OnUpdatePod(msg.Content)
	case message.POD_DELETE:
		ec.OnDeletePod(msg.Content)
	}
}

func (ec *EndpointController) Run() {
	go ec.Consumer.Consume([]string{message.TOPIC_EndpointController}, ec.MsgHandler)
}

func CreateEndpointControllerInstance() (*EndpointController, error) {
	consumer, err := message.NewConsumer(message.TOPIC_EndpointController, "EndpointController")
	return &EndpointController{
		Consumer: consumer,
	}, err
}
