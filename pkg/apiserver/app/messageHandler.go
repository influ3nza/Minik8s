package app

import (
	"encoding/json"
	"fmt"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/message"
	"minik8s/tools"
)

func (s *ApiServer) MsgHandler(msg *message.Message) {
	fmt.Printf("[Apiserver/MsgHandler] Apiserver received message!\n")
	fmt.Printf("[Apiservr/MsgHandler] Message type: %s.\n", msg.Type)

	switch msg.Type {
	case message.POD_CREATE: //kubelet创建完pod或者srv之后，发送消息给apiserver
		s.HandlePodCreate(msg.Content)
	case message.POD_UPDATE:
		s.HandlePodUpdate(msg.Content)
	case message.POD_DELETE:
		s.HandlePodDelete(msg.Content)
	case message.SRV_CREATE:
		s.HandleSrvCreate(msg.Content)
	case message.SRV_DELETE:
		s.HandleSrvDelete(msg.Content)
	}
}

func (s *ApiServer) HandlePodCreate(msg string) {
	pod := &api_obj.Pod{}
	err := json.Unmarshal([]byte(msg), pod)
	if err != nil {
		fmt.Printf("[ERR/Apiserver/MsgHandler/PodCreate] Failed to unmarshal data, %v.\n", err)
		return
	}

	//更新部分信息
	old_str, err := s.UpdatePodPhase(*pod, false)
	if err != nil {
		fmt.Printf("[ERR/Apiserver/MsgHandler/PodCreate] Failed to update pod phase, %v.\n", err)
		return
	}

	//给endpointController发送消息
	p_msg := &message.Message{
		Type:    message.POD_CREATE,
		Content: string(old_str),
	}
	s.Producer.Produce(message.TOPIC_EndpointController, p_msg)
	//WARN:思考如果返回的消息是pod创建失败的消息应该怎么办

	//WARN:仅供测试使用。
	if tools.Test_enabled {
		tools.Pod_created = true
	}
}

func (s *ApiServer) HandlePodUpdate(msg string) {
	//这里只能是Phase改变了。但是如果发现是running且ip地址变化，则需要通知ep controller
	pod := &api_obj.Pod{}
	err := json.Unmarshal([]byte(msg), pod)
	if err != nil {
		fmt.Printf("[ERR/Apiserver/MsgHandler/PodCreate] Failed to unmarshal data, %v.\n", err)
		return
	}

	needRestart := false
	needCheckRestart := false

	//只有phase改变了才会发消息
	switch pod.PodStatus.Phase {
	case obj_inner.Running:
	case obj_inner.Pending:
	case obj_inner.Succeeded:
	case obj_inner.Terminating:
		needRestart = true
	case obj_inner.Failed:
		needRestart = true
	case obj_inner.Unknown:
	case obj_inner.Restarting:
		needCheckRestart = true
	}

	//更新pod状态到etcd
	e_key := apiserver.ETCD_pod_prefix + pod.MetaData.NameSpace + "/" + pod.MetaData.Name
	//如果要创建新的pod则在这里创建。
	if needRestart {
		err = s.EtcdWrap.Del(e_key)
		if err != nil {
			fmt.Printf("[ERR/Apiserver/MsgHandler/PodUpdate] Failed to delete from etcd, %v.\n", err)
			return
		}

		//向ep controller发送删除消息。
		ep_msg := &message.Message{
			Type:    message.POD_DELETE,
			Content: msg,
		}
		s.Producer.Produce(message.TOPIC_EndpointController, ep_msg)

		s.PodNeedRestart(*pod)
	} else {
		_, err := s.UpdatePodPhase(*pod, needCheckRestart)
		if err != nil {
			fmt.Printf("[ERR/Apiserver/MsgHandler/PodUpdate] Failed to update pod phase, %v.\n", err)
			return
		}
	}
	//是否一定要在原node创建？-> 不需要
	//是否可以以同一个名字，同一个配置创建？-> 不可以 -> 可以考虑将挂掉的pod重命名
}

func (s *ApiServer) HandlePodDelete(msg string) {
	//这里默认发送的是"podnamespace/podname"
	//do nothing，因为删除pod的时候apiserver会直接向ep controller发送消息。
}

func (s *ApiServer) HandleSrvCreate(msg string) {
	//将etcd中的srv对象的status修改为available
	//参数：srv结构体
	//TODO:检查一下srv的状态，避免多台机器同时发送相同的消息。
	srv := &api_obj.Service{}
	err := json.Unmarshal([]byte(msg), srv)
	if err != nil {
		fmt.Printf("[ERR/Apiserver/MsgHandler/SrvCreate] Failed to unmarshal data, %v.\n", err)
		return
	}

	err = s.UpdateSrvCondition(srv.MetaData.NameSpace, srv.MetaData.Name)
	if err != nil {
		fmt.Printf("[ERR/Apiserver/MsgHandler/SrvCreate] Failed to update srv condition, %v.\n", err)
		return
	}

	//然后通知endpoint controller
	ep_msg := &message.Message{
		Type:    message.SRV_CREATE,
		Content: msg,
	}
	s.Producer.Produce(message.TOPIC_EndpointController, ep_msg)
}

func (s *ApiServer) HandleSrvDelete(msg string) {
	//kube_proxy会自动将srv下面的所有ep删除。
	//do nothing
}
