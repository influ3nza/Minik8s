package app

import (
	"encoding/json"
	"fmt"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/message"
)

func (s *ApiServer) MsgHandler(msg *message.Message) {
	fmt.Printf("[Apiserver/MsgHandler] Apiserver received message!\n")

	switch msg.Type {
	case message.POD_CREATE: //kubelet创建完pod之后，发送消息给apiserver
		s.HandlePodCreate(msg.Content)
	case message.POD_UPDATE:
		s.HandlePodUpdate(msg.Content)
	case message.POD_DELETE:
		s.HandlePodDelete(msg.Content)
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
	old_str, err := s.UpdatePodPhase(*pod)
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
	//TODO:思考如果返回的消息是pod创建失败的消息应该怎么办
}

func (s *ApiServer) HandlePodUpdate(msg string) {
	//这里只能是Phase改变了。所以不需要给endpointController发送消息。
	pod := &api_obj.Pod{}
	err := json.Unmarshal([]byte(msg), pod)
	if err != nil {
		fmt.Printf("[ERR/Apiserver/MsgHandler/PodCreate] Failed to unmarshal data, %v.\n", err)
		return
	}

	needRestart := false

	switch pod.PodStatus.Phase {
	case obj_inner.Running:
		//break
	case obj_inner.Pending:
		//break
	case obj_inner.Succeeded:
		//break
	case obj_inner.Terminating:
		needRestart = true
	case obj_inner.Failed:
		needRestart = true
	case obj_inner.Unknown:
		//break
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
		s.PodNeedRestart(*pod)
	} else {
		_, err := s.UpdatePodPhase(*pod)
		if err != nil {
			fmt.Printf("[ERR/Apiserver/MsgHandler/PodUpdate] Failed to update pod phase, %v.\n", err)
			return
		}
	}
	//是否一定要在原node创建？-> 不需要
	//是否可以以同一个名字，同一个配置创建？-> 不可以 -> 可以考虑将挂掉的pod重命名
}

func (s *ApiServer) HandlePodDelete(msg string) {
	//TODO:这里默认发送的是"podnamespace/podname"
	//do nothing
}
