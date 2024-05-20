package server

import (
	"fmt"
	"minik8s/pkg/message"
	"minik8s/pkg/serverless/function"
	"minik8s/pkg/serverless/workflow"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type SL_server struct {
	Consumer           *message.MsgConsumer
	FunctionController function.FunctionController
	WorkflowController workflow.WorkflowController
}

func CreateNewSLServerInstance() (*SL_server, error) {
	consumer, err := message.NewConsumer("ss", "ss")
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
	//有关于DELETE应该在apiserver处就被处理完毕了，注意需要同时删除所有pod
	//和对应的rs。
	//UPDATE不考虑。
	}
}

func (s *SL_server) PollApiserver() {
	//TODO:定期轮询apiserver有关func pod的复制数量以及所在ip
	//TODO:需要用某个全局变量存起来。
	//TODO:修改时，需要拿锁。
	for {
		time.Sleep(3 * time.Second)

		pods, err := s.GetPodsFromApiserver()
		if err != nil {
			fmt.Printf("[ERR/Serverless/PollApiserver] Failed to get from apiserver, %v", err)
			return
		}

		s.GeneratePodMap(pods)
	}
}

func (s *SL_server) Run() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	go func() {
		<-sigChan
		s.Clean()
	}()

	//TODO:修改topic
	go s.Consumer.Consume([]string{"ss"}, s.MsgHandler)
	go s.PollApiserver()
}

func (s *SL_server) Clean() {
	fmt.Printf("[apiserver/CLEAN] Apiserver closing...\n")

	close(s.Consumer.Sig)
	s.Consumer.Consumer.Close()
	os.Exit(0)
}
