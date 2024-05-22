package server

import (
	"fmt"
	"minik8s/pkg/message"
	"minik8s/pkg/serverless/function"
	"minik8s/pkg/serverless/workflow"
	"os"
	"os/signal"
	"syscall"
)

type SL_server struct {
	Consumer           *message.MsgConsumer
	FunctionController function.FunctionController
	WorkflowController workflow.WorkflowController
}

func CreateNewSLServerInstance() (*SL_server, error) {
	consumer, err := message.NewConsumer(message.TOPIC_Serverless, message.TOPIC_Serverless)
	if err != nil {
		fmt.Printf("[ERR/Serverless/Server] Error creating consumer.\n")
		return nil, err
	}

	return &SL_server{
		Consumer:           consumer,
		FunctionController: function.CreateNewFunctionControllerInstance(),
		WorkflowController: workflow.CreateNewWorkflowControllerInstance(),
	}, nil
}

func (s *SL_server) MsgHandler(msg *message.Message) {
	switch msg.Type {
	//TODO:填写正确的消息类型：
	//1. FUNC_CREATE
	//2. WF_CREATE
	//3. FUNC_EXEC
	//有关于DELETE应该在apiserver处就被处理完毕了，注意需要同时删除所有pod
	//和对应的rs。
	//UPDATE不考虑。
	}
}

func (s *SL_server) PollApiserver() {
	//TODO:
}

func (s *SL_server) Run() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	go func() {
		<-sigChan
		s.Clean()
	}()

	go s.Consumer.Consume([]string{message.TOPIC_Serverless}, s.MsgHandler)
	go s.PollApiserver()
}

func (s *SL_server) Clean() {
	fmt.Printf("[Serverless/CLEAN] Serverless closing...\n")

	close(s.Consumer.Sig)
	s.Consumer.Consumer.Close()
	os.Exit(0)
}
