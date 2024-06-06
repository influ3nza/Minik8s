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

func (wfc *WorkflowController) ExecuteWorkflow(wf *api_obj.Workflow, fc *function.FunctionController) {
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
		case api_obj.WF_Call:
			//获取wf，执行
			wfc.DoCall(coeff, node, fc)
			pos = node.CallSpec.Next
		case api_obj.WF_Func:
			//执行函数，更新coeff
			coeff, err = wfc.DoFunc(coeff, funcMap[node.FuncSpec.Name], fc)
			if err != nil {
				fmt.Printf("[ERR/WorkflowController] Failed to execute func, %v", err)
			}
			pos = node.FuncSpec.Next
		case api_obj.WF_Fork:
			var err error = nil
			pos, err = wfc.DecideFork(coeff, node)
			if err != nil {
				fmt.Printf("[ERR/WorkflowController] Failed to execute fork, %v", err)
			}
		}
	}
}

func (wfc *WorkflowController) CheckNodeFunc(wf *api_obj.Workflow) (map[string]api_obj.Function, bool) {
	//检查workflow是否合法（是否调用存在的函数）
	//检查workflow调用的函数是否都有pod实例。如果没有需要冷启动。
	funcMap := make(map[string]api_obj.Function)
	uri := apiserver.API_server_prefix + apiserver.API_check_workflow
	wf_str, err := json.Marshal(wf)
	if err != nil {
		fmt.Printf("[CheckNodeFunc] Failed to marshal data.\n")
		return funcMap, false
	}

	err = network.PostRequestAndParse(uri, wf_str, &funcMap)
	if err != nil {
		fmt.Printf("[ERR/CheckNodeFunc] Failed to send POST request, %s\n", err.Error())
		return funcMap, false
	}

	return funcMap, true
}

func (wfc *WorkflowController) DecideFork(coeff string, node api_obj.WorkflowNode) (string, error) {
	//比较
	coeffMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(coeff), &coeffMap)
	if err != nil {
		return "", err
	}

	for _, spec := range node.ForkSpecs {
		fmt.Printf("[FORK] %s\n", spec.CompareBy)
		if spec.CompareBy == api_obj.WF_AllPass {
			return spec.Next, nil
		}

		// 先检查default
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

func (wfc *WorkflowController) DoCompare(v interface{}, by api_obj.CompareType, to string) bool {
	var v_num float64
	var t_int float64
	var err error = nil
	if strings.Count(string(by), api_obj.WF_ByNum) > 0 {
		switch value := v.(type) {
		case int:
			v_num = float64(value)
		case float64:
			v_num = value
		case string:
			v_num, err = strconv.ParseFloat(value, 64)
			if err != nil {
				fmt.Printf("[ERR/DoCompare] v is not a number!\n")
				return false
			}
		}

		t_int, err = strconv.ParseFloat(to, 64)
		if err != nil {
			fmt.Printf("[ERR/DoCompare] t is not a number!\n")
			return false
		}
	}

	switch by {
	case api_obj.WF_ByNumEqual:
		return v_num == t_int
	case api_obj.WF_ByNumGreater:
		return v_num > t_int
	case api_obj.WF_ByNumLess:
		return v_num < t_int
	case api_obj.WF_ByStrEqual:
		return v.(string) == to
	case api_obj.WF_ByStrNotEqual:
		return v.(string) != to
	default:
		return false
	}
}

func (wfc *WorkflowController) DoCall(coeff string, node api_obj.WorkflowNode, fc *function.FunctionController) {
	wfName := node.CallSpec.WfName
	coeffMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(coeff), &coeffMap)
	if err != nil {
		fmt.Printf("[ERR/DoCall] Failed to unmarshal coeff, err: \n", )
		return
	}

	uri := apiserver.API_server_prefix + apiserver.API_get_workflow_prefix + wfName
	wf := &api_obj.Workflow{}
	err = network.GetRequestAndParse(uri, wf)
	if err != nil {
		fmt.Printf("[ERR/DoCall] Failed to send GET request, %v\n", err)
		return
	}

	//TODO:改变wf的初始参数
	newCoeff := make(map[string]interface{})
	for _, c := range node.CallSpec.InheritCoeff {
		newCoeff[c] = coeffMap[c]
	}

	addCoeff := make(map[string]interface{})
	err = json.Unmarshal([]byte(node.CallSpec.NewCoeff), &addCoeff)
	if err != nil {
		fmt.Printf("[ERR/DoCall] Failed to unmarshal coeff.\n")
		return
	}

	for k, v := range addCoeff {
		newCoeff[k] = v
	}

	newCoeff_str, err := json.Marshal(newCoeff)
	if err != nil {
		fmt.Printf("[ERR/DoCall] Failed to marshal coeff.\n")
		return
	}

	wf.Spec.StartCoeff = string(newCoeff_str)

	//开启一个协程
	go wfc.ExecuteWorkflow(wf, fc)
}
