package server

import (
	"encoding/json"
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/message"
	"minik8s/pkg/serverless/function"
	"minik8s/pkg/serverless/workflow"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

type SL_server struct {
	Consumer           *message.MsgConsumer
	Producer           *message.MsgProducer
	Router             *gin.Engine
	FunctionController *function.FunctionController
	WorkflowController workflow.WorkflowController
}

func CreateNewSLServerInstance() (*SL_server, error) {
	consumer, err := message.NewConsumer(message.TOPIC_Serverless, message.TOPIC_Serverless)
	if err != nil {
		fmt.Printf("[ERR/Serverless/Server] Error creating consumer.\n")
		return nil, err
	}

	producer := message.NewProducer()
	if producer == nil {
		fmt.Printf("[ERR/Serverless/Server] Error creating producer.\n")
		return nil, err
	}

	router := gin.Default()
	if router == nil {
		fmt.Printf("[ERR/Serverless/Server] Error creating router.\n")
		return nil, err
	}

	return &SL_server{
		Consumer:           consumer,
		Producer:           producer,
		Router:             router,
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
	//4. FUNC_DEL
	//5. FUNC_UPDATE
	//有关于DELETE应该在apiserver处就被处理完毕了，注意需要同时删除所有pod
	//和对应的rs。
	case message.FUNC_CREATE:
		s.OnFunctionCreate(content)
	case message.FUNC_EXEC:
		s.OnFunctionExec(content)
	case message.WF_EXEC:
		s.OnWorkflowExec(content)
	case message.FUNC_DEL:
		s.OnFunctionDel(content)
	case message.FUNC_UPDATE:
		s.OnFunctionUpdate(content)
	case message.FUNC_EXEC_LOCAL:
		s.OnFunctionExecLocal(content, msg.Backup)
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
	_, err = s.FunctionController.TriggerFunction(f)
	if err != nil {
		fmt.Println("[ERR/serverless/OnFunctionExec] ", err.Error())
	}

	s.FunctionController.UpdateFunction(f.Metadata.Name)

}

func (s *SL_server) OnFunctionExecLocal(content string, cof string) {
	f := &api_obj.Function{}
	err := json.Unmarshal([]byte(content), f)
	if err != nil {
		fmt.Printf("[ERR/serverless/OnFunctionExecLocal] Failed to unmarshal data.\n")
		return
	}

	_, err = s.FunctionController.TriggerFunctionLocal(f, cof)
	if err != nil {
		fmt.Println("[ERR/serverless/OnFunctionExecLocal] ", err.Error())
	}

}

func (s *SL_server) OnFunctionDel(content string) {
	f := &api_obj.Function{}
	err := json.Unmarshal([]byte(content), f)
	if err != nil {
		fmt.Printf("[ERR/serverless/OnFunctionDel] Failed to unmarshal data.\n")
		return
	}
	err = s.FunctionController.DeleteFunction(f)
	if err != nil {
		fmt.Println("[ERR/serverless/OnFunctionDel] ", err.Error())
	}
}

func (s *SL_server) OnFunctionUpdate(content string) {
	f := &api_obj.Function{}
	err := json.Unmarshal([]byte(content), f)
	if err != nil {
		fmt.Printf("[ERR/serverless/OnFunctionUpdate] Failed to unmarshal data.\n")
		return
	}
	err = s.FunctionController.UpdateFunctionBody(f)
	if err != nil {
		fmt.Println("[ERR/serverless/OnFunctionUpdate] ", err.Error())
	}
}

func (s *SL_server) OnWorkflowExec(content string) {
	wf := &api_obj.Workflow{}
	err := json.Unmarshal([]byte(content), wf)
	if err != nil {
		fmt.Printf("[ERR/serverless/OnFunctionExec] Failed to unmarshal data.\n")
		return
	}

	s.WorkflowController.ExecuteWorkflow(wf, s.FunctionController)
}

func (s *SL_server) OnFunctionExecOnServerless(c *gin.Context) {
	name := c.Param("name")
	var req map[string]interface{}
	err := c.ShouldBind(&req)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/serverless/OnFunctionExecOnServerless] Invalid request." + err.Error(),
		})
		return
	}

	reqStr, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/serverless/OnFunctionExecOnServerless] Failed to marshal data.",
		})
		return
	}

	function.RecordMutex.RLock()
	fr := function.RecordMap[name]
	if fr == nil {
		fmt.Printf("[ERR/serverless/OnFunctionExecOnServerless] Function not found.\n")
		c.JSON(404, gin.H{
			"error": "Function not found.",
		})
		function.RecordMutex.RUnlock()
		return
	}
	f_str, err := json.Marshal(*fr.FuncTion)
	function.RecordMutex.RUnlock()

	if err != nil {
		fmt.Printf("[ERR/serverless/OnFunctionExecOnServerless] Failed to marshal data.\n")
		c.JSON(500, gin.H{
			"error": "Failed to marshal data.",
		})
		return
	}

	m_msg := &message.Message{
		Type:    message.FUNC_EXEC_LOCAL,
		Content: string(f_str),
		Backup:  string(reqStr),
	}
	s.Producer.Produce(message.TOPIC_Serverless, m_msg)
	s.FunctionController.UpdateFunction(name)
}

func (s *SL_server) bindHandler() {
	s.Router.POST("/functions/exec/:name", s.OnFunctionExecOnServerless)
}

func (s *SL_server) Run() {
	s.bindHandler()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	go func() {
		<-sigChan
		s.Clean()
	}()

	// go s.FunctionController.RunWatch()
	go function.Watcher.FileWatch()
	go s.Consumer.Consume([]string{message.TOPIC_Serverless}, s.MsgHandler)
	s.Router.Run(":50001")
}

func (s *SL_server) Clean() {
	fmt.Printf("[Serverless/CLEAN] Serverless closing...\n")

	close(s.Consumer.Sig)
	s.Consumer.Consumer.Close()
	os.Exit(0)
}
