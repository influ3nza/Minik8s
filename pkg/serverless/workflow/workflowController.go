package workflow

import (
	"minik8s/pkg/api_obj"
	"time"
)

type WorkflowController struct {
}

func CreateNewWorkflowControllerInstance() WorkflowController {
	return WorkflowController{}
}

func (wfc *WorkflowController) ExecuteWorkflow(wf api_obj.Workflow) {
	//TODO:该函数可以通过协程调用。（也可以不考虑workflow的并发。）
	//检查workflow是否合法（是否调用存在的函数）
	//检查workflow调用的函数是否都有pod实例。如果没有需要冷启动。
	//按照DAG的流程依次执行相应的function。
	//执行完毕之后，将workflow的状态保存在etcd中。
	for {
		flag := true
		for _, node := range wf.Spec.Nodes {
			if !wfc.CheckNodeFunc(node) {
				flag = false
				break
			}
		}

		if !flag {
			time.Sleep(1 * time.Second)
			continue
		}

		wfc.executeWorkflow(wf)
	}
}

func (wfc *WorkflowController) CheckNodeFunc(node api_obj.WorkflowNode) bool {
	//检查workflow是否合法（是否调用存在的函数）
	//检查workflow调用的函数是否都有pod实例。如果没有需要冷启动。
	
}

func (wfc *WorkflowController) executeWorkflow(wf api_obj.Workflow) {
	pos := wf.Spec.StartNode
	coeff := wf.Spec.StartCoeff

	for {
		if pos == "" {
			break
		}

		node := 
	}
}
