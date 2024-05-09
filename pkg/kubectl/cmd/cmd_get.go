package cmd

import (
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/network"
	"strings"

	"github.com/spf13/cobra"
)

var GetCmd = &cobra.Command{
	Use:     "get <apiobject> [<namespace>][/<name>]",
	Short:   "gets the information of an apiobject",
	Long:    "This is a command for user to check the information of an apiobject. Simply use \"kubectl get <apiobject> <namespace>/<name>\".",
	Example: "kubectl get pod testing/pod1",
	Args:    cobra.MinimumNArgs(1),
	Run:     GetHandler,
}

func GetHandler(cmd *cobra.Command, args []string) {
	apitype := args[0]
	namespace, name := "", ""

	if len(args) == 2 {
		//两个参数 -> 分两种情况
		//单namespace 或者 namespace/name
		key := args[1]
		if strings.Count(key, "/") == 0 {
			namespace = key
			name = ""
		} else if strings.Count(key, "/") == 1 {
			index := strings.Index(key, "/") + 1
			namespace = key[0 : index-1]
			name = key[index:]

			if len(namespace) == 0 || len(name) == 0 {
				fmt.Println("[ERR] Empty namespace or name. Try -h for help.")
				return
			}
		} else {
			fmt.Println("[ERR] Wrong format. Try -h for help.")
			return
		}
	} else if len(args) > 2 {
		fmt.Println("[ERR] Too many arguments. Try -h for help.")
		return
	}

	switch apitype {
	case "pod":
		GetPodHandler(namespace, name)
	case "node":
		GetNodeHandler(namespace, name)
	case "service":
		GetServiceHandler(namespace, name)
	case "replicaset":
		GetReplicasetHandler(namespace, name)
	case "hpa":
		GetHpaHandler(namespace, name)
	default:
		fmt.Println("[ERR] Wrong api kind. Available: pod, node, service, replicaset, hpa.")
	}
}

func GetPodHandler(namespace string, name string) {
	//获取pod(范围取决于两个参数)，整理之后输出必要的信息。
	//NAME  READY  STATUS  RESTARTS  AGE
	uri := ""
	pods := []api_obj.Pod{}
	if namespace == "" && name == "" {
		uri = config.API_server_prefix + config.API_get_pods
		err := network.GetRequestAndParse(uri, &pods)
		if err != nil {
			fmt.Printf("[ERR/GetPodHandler] %v\n", err)
		}
	} else if name == "" {
		uri = config.API_server_prefix + config.API_get_pods_by_namespace_prefix + namespace
	} else {
		uri = config.API_server_prefix +
	}
}

func GetNodeHandler(namespace string, name string) {

}

func GetServiceHandler(namespace string, name string) {

}

func GetReplicasetHandler(namespace string, name string) {

}

func GetHpaHandler(namespace string, name string) {

}

func PrintPodHandler(pods []api_obj.Pod) {
	//打印相关信息
}
