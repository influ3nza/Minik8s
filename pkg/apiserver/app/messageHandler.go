package app

import (
	"fmt"

	"minik8s/pkg/message"
)

func (s *ApiServer) MsgHandler(msg *message.Message) {
	fmt.Printf("[Apiserver/MsgHandler] Apiserver received message!\n")

	switch msg.Type {
	case message.POD_CREATE: //kubelet创建完pod之后，发送消息给apiserver
		s.HandlePodCreate(msg)
	default:
		return
	}
}

func (s *ApiServer) HandlePodCreate(msg *message.Message) {
	//TODO:更新完ip地址以及相关信息之后，向endpointController发送podCreate消息
	//m := &message.Message{}
	//s.Producer.Produce(message.TOPIC_EndpointController, m)

	//TODO:思考如果返回的消息是pod创建失败的消息应该怎么办
}
