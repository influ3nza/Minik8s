package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/network"

	"github.com/olekukonko/tablewriter"
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
		if name != "" && namespace != "" {
			fmt.Println("[ERR] Too many arguments to get node. Try -h for help.")
			return
		}
		GetNodeHandler(namespace)
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
	uri := apiserver.API_server_prefix
	pods := []api_obj.Pod{}
	if namespace == "" && name == "" {
		uri += apiserver.API_get_pods
	} else if name == "" {
		uri += apiserver.API_get_pods_by_namespace_prefix + namespace
	} else {
		uri += apiserver.API_get_pod_prefix + namespace + "/" + name
	}

	err := network.GetRequestAndParse(uri, &pods)
	if err != nil {
		fmt.Printf("[ERR/GetPodHandler] %v\n", err)
		return
	}

	PrintPodHandler(pods)
}

func GetNodeHandler(name string) {
	//node没有namespace！直接传name作为参数
	uri := apiserver.API_server_prefix
	nodes := []api_obj.Node{}
	if name == "" {
		uri += apiserver.API_get_nodes
	} else {
		uri += apiserver.API_get_node_prefix + name
	}

	err := network.GetRequestAndParse(uri, &nodes)
	if err != nil {
		fmt.Printf("[ERR/GetNodeHandler] %v\n", err)
		return
	}

	PrintNodeHandler(nodes)
}

func GetServiceHandler(namespace string, name string) {
	uri := apiserver.API_server_prefix
	srvs := []api_obj.Service{}
	if namespace == "" && name == "" {
		uri += apiserver.API_get_services
	} else if name == "" {
		fmt.Printf("[ERR/GetServiceHandler] Get services by namespace is not supported.\n")
		return
	} else {
		uri += apiserver.API_get_service_prefix + namespace + "/" + name
	}

	err := network.GetRequestAndParse(uri, &srvs)
	if err != nil {
		fmt.Printf("[ERR/GetServiceHandler] %v\n", err)
		return
	}

	PrintServiceHandler(srvs)
}

func GetReplicasetHandler(namespace string, name string) {
	uri := apiserver.API_server_prefix
	rps := []api_obj.ReplicaSet{}
	if namespace != "" || name != "" {
		fmt.Printf("[ERR/GetReplicasetHandler] Get specific replicaset is not supported.\n")
		return
	}

	uri += apiserver.API_get_replicasets
	err := network.GetRequestAndParse(uri, &rps)
	if err != nil {
		fmt.Printf("[ERR/GetReplicasetHandler] %v\n", err)
		return
	}

	PrintReplicasetHandler(rps)
}

func GetHpaHandler(namespace string, name string) {

}

func PrintPodHandler(pods []api_obj.Pod) {
	//打印相关信息
	layout := "2006-01-02 15:04:05"
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAMESPACE", "NAME", "STATUS", "RESTARTS", "AGE"})
	for _, pod := range pods {
		up_time, _ := time.Parse(layout, pod.PodStatus.CreateTime)
		now_time := time.Now()
		delta := now_time.Sub(up_time)
		table.Append([]string{
			pod.MetaData.NameSpace,
			pod.MetaData.Name,
			pod.PodStatus.Phase,
			strconv.Itoa(int(pod.PodStatus.Restarts)),
			delta.String()})
	}
	table.Render()
}

func PrintNodeHandler(nodes []api_obj.Node) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "STATUS", "ROLE"})
	for _, node := range nodes {
		table.Append([]string{node.NodeMetadata.Name, string(node.NodeStatus.Condition), "WORKER"})
	}
	table.Render()
}

func PrintServiceHandler(srvs []api_obj.Service) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAMESPACE", "NAME", "TYPE", "CLUSTERIP"})
	for _, srv := range srvs {
		table.Append([]string{
			srv.MetaData.NameSpace,
			srv.MetaData.Name,
			string(srv.Spec.Type),
			srv.Spec.ClusterIP})
	}
	table.Render()
}

func PrintReplicasetHandler(rps []api_obj.ReplicaSet) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAMESPACE", "NAME", "READY/REPLICAS"})
	for _, rp := range rps {
		table.Append([]string{
			rp.MetaData.NameSpace,
			rp.MetaData.Name,
			strconv.Itoa(rp.Status.ReadyReplicas) + "/" + strconv.Itoa(rp.Spec.Replicas)})
	}
	table.Render()
}

func PrintHpaHandler() {

}
