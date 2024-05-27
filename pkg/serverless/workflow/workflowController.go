package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/network"
	"minik8s/pkg/serverless/function"
	"strconv"
	"strings"
)

type WorkflowController struct {
}

func CreateNewWorkflowControllerInstance() WorkflowController {
	return WorkflowController{}
}

func (wfc *WorkflowController) ExecuteWorkflow(wf api_obj.Workflow, fc *function.FunctionController) {
	//TODO:该函数可以通过协程调用。（也可以不考虑workflow的并发。）
	//检查workflow是否合法（是否调用存在的函数）
	//检查workflow调用的函数是否都有pod实例。如果没有需要冷启动。
	//按照DAG的流程依次执行相应的function。
	//执行完毕之后，将workflow的状态保存在etcd中。
	funcMap, flag := wfc.CheckNodeFunc(wf)
	if !flag {
		fmt.Printf("[ERR/ExecuteWorkflow] Check workflow failed. Quitting.\n")
		return
	}

	pos := wf.Spec.StartNode
	coeff := wf.Spec.StartCoeff
	var err error = nil
	nodesIndex := make(map[string]api_obj.WorkflowNode)

	for _, node := range wf.Spec.Nodes {
		nodesIndex[node.Name] = node
	}

	for {
		if pos == "" {
			break
		}

		node := nodesIndex[pos]
		switch node.Type {
		case api_obj.WF_Func:
			//执行函数，更新coeff
			coeff, err = wfc.DoFunc(coeff, funcMap[node.FuncSpec.Name], fc)
			if err != nil {
				fmt.Printf("[ERR/WorkflowController] Failed to execute func, %v", err)
			}
		case api_obj.WF_Fork:
			var err error = nil
			pos, err = wfc.DecideFork(coeff, node)
			if err != nil {
				fmt.Printf("[ERR/WorkflowController] Failed to execute fork, %v", err)
			}
		}
	}
}

func (wfc *WorkflowController) CheckNodeFunc(wf api_obj.Workflow) (map[string]api_obj.Function, bool) {
	//检查workflow是否合法（是否调用存在的函数）
	//检查workflow调用的函数是否都有pod实例。如果没有需要冷启动。
	funcMap := make(map[string]api_obj.Function)
	uri := apiserver.API_server_prefix + apiserver.API_check_workflow
	err := network.GetRequestAndParse(uri, &funcMap)
	if err != nil {
		fmt.Printf("[ERR/CheckNodeFunc] Failed to send GET request.\n")
		return funcMap, false
	}

	return funcMap, true
}

func (wfc *WorkflowController) DecideFork(coeff string, node api_obj.WorkflowNode) (string, error) {
	//比较
	coeffMap := make(map[string]string)
	err := json.Unmarshal([]byte(coeff), &coeffMap)
	if err != nil {
		return "", err
	}

	for _, spec := range node.ForkSpecs {
		v := coeffMap[spec.Variable]
		if wfc.DoCompare(v, spec.CompareBy, spec.CompareTo) {
			return spec.Next, nil
		}
	}

	return "", errors.New("shall not reach here")
}

func (wfc *WorkflowController) DoFunc(coeff string, f api_obj.Function, fc *function.FunctionController) (string, error) {
	f.Coeff = coeff
	res, err := fc.TriggerFunction(&f)
	return res, err
}

func (wfc *WorkflowController) DoCompare(v string, by api_obj.CompareType, to string) bool {
	v_int := 0
	t_int := 0
	var err error = nil
	if strings.Count(string(by), api_obj.WF_ByNum) > 0 {
		v_int, err = strconv.Atoi(v)
		if err != nil {
			fmt.Printf("[ERR/DoCompare] v is not a number!\n")
			return false
		}

		t_int, err = strconv.Atoi(to)
		if err != nil {
			fmt.Printf("[ERR/DoCompare] t is not a number!\n")
			return false
		}
	}

	switch by {
	case api_obj.WF_ByNumEqual:
		return v_int == t_int
	case api_obj.WF_ByNumGreater:
		return v_int > t_int
	case api_obj.WF_ByNumLess:
		return v_int < t_int
	case api_obj.WF_ByStrEqual:
		return v == to
	case api_obj.WF_ByStrNotEqual:
		return v != to
	default:
		return false
	}
}
