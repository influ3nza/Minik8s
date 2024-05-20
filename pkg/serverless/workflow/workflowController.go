package workflow

import (
	"minik8s/pkg/api_obj"
)

type WorkflowController struct {
	//TODO:可能需要消息队列？
}

func CreateNewWorkflowControllerInstance() WorkflowController {
	return WorkflowController{}
}

func (wfc *WorkflowController) ExecuteWorkflow(wf api_obj.Workflow) {
	//TODO:该函数可以通过协程调用。（也可以不考虑workflow的并发。）
	//TODO:检查workflow是否合法（是否调用存在的函数）
	//TODO:检查workflow调用的函数是否都有pod实例。如果没有需要冷启动。
	//TODO:按照DAG的流程依次执行相应的function。
	//TODO:执行完毕之后，将workflow的状态保存在etcd中。
}
