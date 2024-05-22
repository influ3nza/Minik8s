package cmd

import (
	"fmt"

	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/network"

	"github.com/spf13/cobra"
)

var ExecCmd = &cobra.Command{
	Use:     "exec",
	Short:   "kubectl executes an serverless function",
	Long:    "This is a command for user to execute serverless functions. Simply use \"kubectl exec <function_name> [{<coeff>}]\".",
	Example: "kubectl exec add_sum {a = 1, b = 2}",
	Args:    cobra.MinimumNArgs(1),
	Run:     ExecHandler,
}

func ExecHandler(cmd *cobra.Command, args []string) {
	func_name := args[0]
	coeff := "{}"
	if len(args) == 2 {
		coeff = args[1]
	} else if len(args) > 2 {
		fmt.Printf("[ERR/ExecHandler] Too many arguments. Try -h for help.\n")
		return
	}
	//向apiserver发送函数执行请求
	uri := apiserver.API_server_prefix + apiserver.API_exec_function_prefix + func_name + "/" + coeff
	data, err := network.GetRequest(uri)
	if err != nil {
		fmt.Printf("[ERR/ExecHandler] Failed to send GET request, %v", err)
		return
	}

	fmt.Printf("[ExecHandler] Function execute success, return value: %s", data)
}
