package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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
		GetReplicasetHandler()
	case "hpa":
		GetHpaHandler()
	case "function":
		GetFunctionHandler()
	case "dns":
		GetDnsHandler()
	case "workflow":
		GetWorkFlowHandler()
	case "pv":
		if name != "" && namespace != "" {
			fmt.Println("[ERR] Too many arguments to get pv. Try -h for help.")
			return
		}
		GetPVHandler()
	case "pvc":
		if name != "" && namespace != "" {
			fmt.Println("[ERR] Too many arguments to get pvc. Try -h for help.")
			return
		}
		GetPVCHandler()
	default:
		fmt.Println("[ERR] Wrong api kind. Available: pod, node, service, replicaset, hpa, dns, function, workflow.")
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

func GetReplicasetHandler() {
	uri := apiserver.API_server_prefix
	rps := []api_obj.ReplicaSet{}

	uri += apiserver.API_get_replicasets
	err := network.GetRequestAndParse(uri, &rps)
	if err != nil {
		fmt.Printf("[ERR/GetReplicasetHandler] %v\n", err)
		return
	}

	PrintReplicasetHandler(rps)
}

func GetDnsHandler() {
	uri := apiserver.API_server_prefix + apiserver.API_get_all_dns
	dss := []api_obj.Dns{}
	err := network.GetRequestAndParse(uri, &dss)
	if err != nil {
		fmt.Printf("[ERR/GetDnsHandler] %v\n", err)
		return
	}

	PrintDnsHandler(dss)
}

func GetHpaHandler() {
	uri := apiserver.API_server_prefix + apiserver.API_get_hpas
	hpas := []api_obj.HPA{}
	err := network.GetRequestAndParse(uri, &hpas)
	if err != nil {
		fmt.Printf("[ERR/GetHpaHandler] %v\n", err)
		return
	}

	PrintHpaHandler(hpas)
}

func GetFunctionHandler() {
	uri := apiserver.API_server_prefix + apiserver.API_get_all_functions
	fs := []api_obj.Function{}
	err := network.GetRequestAndParse(uri, &fs)
	if err != nil {
		fmt.Printf("[ERR/GetFunctionHandler] %v\n", err)
		return
	}

	PrintFunctionHandler(fs)
}

func GetWorkFlowHandler() {
	uri := apiserver.API_server_prefix + apiserver.API_get_all_workflows
	wfs := []api_obj.Workflow{}
	err := network.GetRequestAndParse(uri, &wfs)
	if err != nil {
		fmt.Printf("[ERR/GetWorkflowHandler] %v\n", err)
		return
	}

	PrintWorkflowHandler(wfs)
}

func GetPVHandler() {
	uri := apiserver.API_server_prefix
	pvs := []api_obj.PV{}

	uri += apiserver.API_get_pvs
	err := network.GetRequestAndParse(uri, &pvs)
	if err != nil {
		fmt.Printf("[ERR/GetPVHandler] %v\n", err)
		return
	}

	PrintPVHandler(pvs)
}

func GetPVCHandler() {
	uri := apiserver.API_server_prefix
	pvcs := []api_obj.PVC{}

	uri += apiserver.API_get_pvcs
	err := network.GetRequestAndParse(uri, &pvcs)
	if err != nil {
		fmt.Printf("[ERR/GetPVCHandler] %v\n", err)
		return
	}

	PrintPVCHandler(pvcs)
}

func PrintPodHandler(pods []api_obj.Pod) {
	//打印相关信息
	// // layout := "2006-01-02 15:04:05"
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAMESPACE", "NAME", "STATUS", "RESTARTS", "NODE", "IP"})
	for _, pod := range pods {
		// // up_time, _ := time.Parse(layout, pod.PodStatus.CreateTime)
		// // now_time := time.Now()
		// // delta := now_time.Sub(up_time)
		table.Append([]string{
			pod.MetaData.NameSpace,
			pod.MetaData.Name,
			pod.PodStatus.Phase,
			strconv.Itoa(int(pod.PodStatus.Restarts)),
			// // delta.String(),
			pod.Spec.NodeName,
			pod.PodStatus.PodIP,
			pod.PodStatus.PodIP,
		})
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
	table.SetHeader([]string{"NAMESPACE", "NAME", "SELECTOR", "CLUSTERIP", "PORT/TARGETPORT/NODEPORT", "ENDPOINTS"})
	for _, srv := range srvs {
		selector := ""
		for k, v := range srv.Spec.Selector {
			selector += k + ": " + v + "\n"
		}
		if selector != "" {
			selector = selector[:len(selector)-1]
		}

		uri := apiserver.API_server_prefix + apiserver.API_get_endpoint_by_service_prefix + srv.MetaData.NameSpace + "/" + srv.MetaData.Name
		eps := []api_obj.Endpoint{}
		err := network.GetRequestAndParse(uri, &eps)
		if err != nil {
			fmt.Printf("[ERR/kubectl/GetSrvs] Failed to send GET request, %s.\n", err.Error())
			return
		}

		epip := ""
		for _, ep := range eps {
			epip += ep.PodIP + "\n"
		}
		if epip != "" {
			epip = epip[:len(epip)-1]
		}

		table.Append([]string{
			srv.MetaData.NameSpace,
			srv.MetaData.Name,
			selector,
			srv.Spec.ClusterIP,
			strconv.Itoa(int(srv.Spec.Ports[0].Port)) + "/" + strconv.Itoa(int(srv.Spec.Ports[0].TargetPort)) + "/" + strconv.Itoa(int(srv.Spec.Ports[0].NodePort)),
			epip,
		})
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

func PrintPVHandler(pvs []api_obj.PV) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "SERVERIP", "PATH"})
	for _, pv := range pvs {
		table.Append([]string{
			pv.Metadata.Name,
			pv.Spec.Nfs.ServerIp,
			pv.Spec.Nfs.Path,
		})
	}
	table.Render()
}

func PrintPVCHandler(pvcs []api_obj.PVC) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "BINDPV"})
	for _, pvc := range pvcs {
		table.Append([]string{
			pvc.Metadata.Name,
			pvc.Spec.BindPV,
		})
	}
	table.Render()
}

func PrintHpaHandler(hpas []api_obj.HPA) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAMESPACE", "NAME", "MIN/NOW/MAX REPLICAS"})
	for _, hpa := range hpas {
		table.Append([]string{
			hpa.MetaData.NameSpace,
			hpa.MetaData.Name,
			strconv.Itoa(hpa.Spec.MinReplicas) + "/" + strconv.Itoa(hpa.Status.CurReplicas) + "/" + strconv.Itoa(hpa.Spec.MaxReplicas),
		})
	}
	table.Render()
}

func PrintDnsHandler(dss []api_obj.Dns) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "HOST", "PATH", "SRVNAME"})
	for _, dns := range dss {
		path := ""
		srvname := ""
		for _, p := range dns.Paths {
			path += "/" + p.SubPath + "\n"
			srvname += p.ServiceName + "\n"
		}
		if path != "" {
			path = path[:len(path)-1]
		}
		if srvname != "" {
			srvname = srvname[:len(srvname)-1]
		}

		table.Append([]string{
			dns.MetaData.Name,
			dns.Host,
			path,
			srvname,
		})
	}
	table.Render()
}

func PrintFunctionHandler(fs []api_obj.Function) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "FILEPATH", "NEEDWATCH"})
	for _, f := range fs {
		table.Append([]string{
			f.Metadata.Name,
			f.FilePath,
			fmt.Sprintf("%t", f.NeedWatch),
		})
	}
	table.Render()
}

func PrintWorkflowHandler(wfs []api_obj.Workflow) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME"})
	for _, wf := range wfs {
		table.Append([]string{
			wf.MetaData.Name,
		})
	}
	table.Render()
}
