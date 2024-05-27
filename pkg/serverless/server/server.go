package server

import (
	"encoding/json"
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/message"
	"minik8s/pkg/serverless/function"
	"minik8s/pkg/serverless/workflow"
	"os"
	"os/signal"
	"syscall"
)

type SL_server struct {
	Consumer           *message.MsgConsumer
	Producer           *message.MsgProducer
	FunctionController *function.FunctionController
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
	fmt.Printf("[SL_server/MsgHandler] Received message!\n")
	content := msg.Content

	switch msg.Type {
	//填写正确的消息类型：
	//1. FUNC_CREATE
	//2. WF_CREATE
	//3. FUNC_EXEC
	//有关于DELETE应该在apiserver处就被处理完毕了，注意需要同时删除所有pod
	//和对应的rs。
	//UPDATE不考虑。
	case message.FUNC_CREATE:
		s.OnFunctionCreate(content)
	case message.FUNC_EXEC:
		s.OnFunctionExec(content)
	case message.WF_EXEC:
		s.OnWorkflowExec(content)
	case message.FUNC_DEL:
		s.OnFunctionDel(content)
	}
}

func (s *SL_server) OnFunctionCreate(content string) {
	f := &api_obj.Function{}
	err := json.Unmarshal([]byte(content), f)
	if err != nil {
		fmt.Printf("[ERR/serverless/OnFunctionCreate] Failed to unmarshal data.\n")
		return
	}

	err = s.FunctionController.GenerateFunction(f)
	if err != nil {
		fmt.Printf("[ERR/serverless/OnFunctionCreate] Failed to Create, %s", err.Error())
		return
	}
}

func (s *SL_server) OnFunctionExec(content string) {
	f := &api_obj.Function{}
	err := json.Unmarshal([]byte(content), f)
	if err != nil {
		fmt.Printf("[ERR/serverless/OnFunctionExec] Failed to unmarshal data.\n")
		return
	}
	s.FunctionController.TriggerFunction(f)
}

func (s *SL_server) OnFunctionDel(content string) {
	f := &api_obj.Function{}
	err := json.Unmarshal([]byte(content), f)
	if err != nil {
		fmt.Printf("[ERR/serverless/OnFunctionDel] Failed to unmarshal data.\n")
		return
	}
	s.FunctionController.DeleteFunction(f)
}

func (s *SL_server) OnWorkflowExec(content string) {
	wf := &api_obj.Workflow{}
	err := json.Unmarshal([]byte(content), wf)
	if err != nil {
		fmt.Printf("[ERR/serverless/OnFunctionExec] Failed to unmarshal data.\n")
		return
	}

	s.WorkflowController.ExecuteWorkflow(*wf, s.FunctionController)
}

func (s *SL_server) Run() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	go func() {
		<-sigChan
		s.Clean()
	}()

	go s.Consumer.Consume([]string{message.TOPIC_Serverless}, s.MsgHandler)
}

func (s *SL_server) Clean() {
	fmt.Printf("[Serverless/CLEAN] Serverless closing...\n")

	close(s.Consumer.Sig)
	s.Consumer.Consumer.Close()
	os.Exit(0)
}
