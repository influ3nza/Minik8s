package cmd

import (
	"fmt"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/network"
	"os"

	"github.com/spf13/cobra"
)

var showResult bool = false
var useFileCoeff bool = false

var ExecCmd = &cobra.Command{
	Use:     "exec",
	Short:   "kubectl executes an serverless function or workflow",
	Long:    "This is a command for user to execute serverless functions or workflows. Simply use \"kubectl exec function/workflow [-r] <function_name> [{<coeff>}]\".",
	Example: "kubectl exec function -r add_sum {a = 1, b = 2}",
	Args:    cobra.MinimumNArgs(1),
	Run:     ExecHandler,
}

func init() {
	ExecCmd.PersistentFlags().BoolVarP(&showResult, "result", "r", false, "Show function result")
	ExecCmd.PersistentFlags().BoolVarP(&useFileCoeff, "file", "f", false, "Import file as coeff")
}

func ExecHandler(cmd *cobra.Command, args []string) {
	switch args[0] {
	case "function":
		FuncExecHandler(args)
	case "workflow":
		WorkflowExecHandler(args)
	}
}

func FuncExecHandler(args []string) {
	func_name := args[1]
	coeff := "{}"
	fmt.Println(args, len(args))
	if len(args) == 3 {
		coeff = args[2]
	} else if len(args) > 3 {
		fmt.Printf("[ERR/ExecHandler] Too many arguments. Try -h for help.\n")
		return
	}

	if !showResult {
		// 向apiserver发送函数执行请求
		uri := apiserver.API_server_prefix + apiserver.API_exec_function_prefix + func_name
		_, err := network.PostRequest(uri, []byte(coeff))
		if err != nil {
			fmt.Printf("[ERR/ExecHandler] Failed to send GET request, %v", err)
			return
		}
	} else {
		//向apiserver请求函数执行结果。

	}
}

func WorkflowExecHandler(args []string) {
	wf_name := args[1]
	coeff := "{}"

	if len(args) == 3 {
		if !useFileCoeff {
			coeff = args[2]
		} else {
			fileBytes, err := os.ReadFile(args[2])
			if err != nil {
				fmt.Printf("[ERR/ExecHandler] Failed to read file, %s.\n", err.Error())
				return
			}
			coeff = string(fileBytes)
		}
	} else if len(args) > 3 {
		fmt.Printf("[ERR/ExecHandler] Too many arguments. Try -h for help.\n")
		return
	}

	if !showResult {
		//向apiserver发送函数执行请求
		uri := apiserver.API_server_prefix + apiserver.API_exec_workflow_prefix + wf_name
		_, err := network.PostRequest(uri, []byte(coeff))
		if err != nil {
			fmt.Printf("[ERR/ExecHandler] Failed to send GET request, %v", err)
			return
		}
	} else {
		//向apiserver请求函数执行结果。

	}
}
