package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/message"
	"minik8s/pkg/network"
	"minik8s/pkg/serverless/function"
	"strconv"
	"strings"
)

type WorkflowController struct {
}

type WorkflowExecutor struct {
	MyTopic    string
	BlockTopic string
	Consumer   *message.MsgConsumer
	Producer   *message.MsgProducer
	Wf         *api_obj.Workflow
}

func CreateNewWorkflowControllerInstance() WorkflowController {
	return WorkflowController{}
}

func CreateNewWorkflowExecutorInstance(wf *api_obj.Workflow, myTopic string, blockTopic string) (*WorkflowExecutor, error) {
	producer := message.NewProducer()
	consumer, err := message.NewConsumer(blockTopic, blockTopic)
	if err != nil {
		fmt.Printf("[ERR/WorkflowExecutor] create kafka consumer instance failed, err:%v\n", err)
		return nil, err
	}

	return &WorkflowExecutor{
		MyTopic:    myTopic,
		BlockTopic: blockTopic,
		Consumer:   consumer,
		Producer:   producer,
		Wf:         wf,
	}, nil
}

func (wfc *WorkflowController) ExecuteWorkflow(wf *api_obj.Workflow, fc *function.FunctionController) {
	//TODO:该函数可以通过协程调用。（也可以不考虑workflow的并发。）
	//检查workflow是否合法（是否调用存在的函数）
	//检查workflow调用的函数是否都有pod实例。如果没有需要冷启动。
	//按照DAG的流程依次执行相应的function。
	//执行完毕之后，将workflow的状态保存在etcd中。
	exec, f := CreateNewWorkflowExecutorInstance(wf, wf.MetaData.Name, caller)
	if f != nil {
		fmt.Printf("[ERR/Execworkflow] Failed to createworkflow executor, %s\n", f.Error())
		return
	}
	
	exec.ExecuteWorkflow(wf, fc, "")
}

func (exec *WorkflowExecutor) ExecuteWorkflow(wf *api_obj.Workflow, fc *function.FunctionController, caller string) {
	funcMap, flag := exec.CheckNodeFunc(wf)
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
			//TODO:获取wf，执行，阻塞
			exec.DoCall()
			pod = node.CallSpec.Next
		case api_obj.WF_Func:
			//执行函数，更新coeff
			coeff, err = exec.DoFunc(coeff, funcMap[node.FuncSpec.Name], fc)
			if err != nil {
				fmt.Printf("[ERR/WorkflowController] Failed to execute func, %v", err)
			}
			pos = node.FuncSpec.Next
		case api_obj.WF_Fork:
			var err error = nil
			pos, err = exec.DecideFork(coeff, node)
			if err != nil {
				fmt.Printf("[ERR/WorkflowController] Failed to execute fork, %v", err)
			}
		}
	}

	//TODO:将结果打印到某个文件中
}

func (exec *WorkflowExecutor) CheckNodeFunc(wf *api_obj.Workflow) (map[string]api_obj.Function, bool) {
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

func (exec *WorkflowExecutor) DecideFork(coeff string, node api_obj.WorkflowNode) (string, error) {
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
		if exec.DoCompare(v, spec.CompareBy, spec.CompareTo) {
			return spec.Next, nil
		}
	}

	return "", errors.New("shall not reach here")
}

func (exec *WorkflowExecutor) DoFunc(coeff string, f api_obj.Function, fc *function.FunctionController) (string, error) {
	f.Coeff = coeff
	res, err := fc.TriggerFunction(&f)
	return res, err
}

func (exec *WorkflowExecutor) DoCompare(v interface{}, by api_obj.CompareType, to string) bool {
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

func (exec *WorkflowExecutor) DoCall(node api_obj.WorkflowNode) {
	wfName := node.CallSpec.WfName

	uri := 
}