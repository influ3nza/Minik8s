package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/config/kube_proxy"
	"minik8s/pkg/network"
	"minik8s/tools"
	"strconv"
)

func CompareLabels(a map[string]string, b map[string]string) bool {
	for key, val := range a {
		if b[key] != val {
			return false
		}
	}
	return true
}

func CreateEndpoints(srvs []api_obj.Service, pods []api_obj.Pod) error {
	//在应用场景中，可以确保srvs和pods至少有一个长度仅为1。
	ep_list := []api_obj.Endpoint{}

	for _, srv := range srvs {
		for _, pod := range pods {
			for _, pair := range srv.Spec.Ports {
				port := GetMatchPort(pair.TargetPort, pod.Spec.Containers)
				if port < 0 {
					return errors.New("no matching port")
				}

				ep := &api_obj.Endpoint{
					MetaData: obj_inner.ObjectMeta{
						Name:      srv.MetaData.Name + "-" + pod.MetaData.Name,
						NameSpace: srv.MetaData.NameSpace,
					},
					SrvIP:   srv.Spec.ClusterIP,
					SrvPort: srv.Spec.Ports[0].Port,
					PodUUID: pod.MetaData.UUID,
					PodIP:   pod.PodStatus.PodIP,
					PodPort: port,
				}

				ep_list = append(ep_list, *ep)

				ep_str, err := json.Marshal(ep)
				if err != nil {
					fmt.Printf("[ERR/Controller/Utils/Endpoint] Failed to marshal endpoint, " + err.Error())
					return err
				}

				fmt.Printf("[Controller/Utils/Endpoint] Try to create endpoint %s.\n", ep.MetaData.Name)

				uri := apiserver.API_server_prefix + apiserver.API_add_endpoint
				_, err = network.PostRequest(uri, ep_str)
				if err != nil {
					fmt.Printf("[ERR/Controller/Utils/Endpoint] GET request failed, %v.\n", err)
					return err
				}

				fmt.Printf("[Controller/Utils/Endpoint] Endpoint create request for srv %s:%d, pod %s:%d.\n",
					srv.Spec.ClusterIP, pair.TargetPort, pod.PodStatus.PodIP, port)
			}
		}
	}

	if len(ep_list) == 0 {
		fmt.Printf("[Controller/Utils/Endpoint] No endpoints to create, return.\n")
		return nil
	}

	//以数组的形式向proxy发送创建请求。
	ep_list_str, err := json.Marshal(ep_list)
	if err != nil {
		fmt.Printf("[ERR/Controller/Utils/Endpoint] Failed to marshal data, %v.\n", err)
		return err
	}

	for _, ip := range tools.NodesIpMap {
		uri := ip + strconv.Itoa(int(kube_proxy.Port)) + kube_proxy.AddEndpoint
		_, err = network.PostRequest(uri, ep_list_str)
		if err != nil {
			fmt.Printf("[ERR/Controller/Utils/Endpoint] Failed to send POST request, %v.\n", err)
			return err
		}
	}

	return nil
}

func GetMatchPort(srvPort int32, cons []api_obj.Container) int32 {
	//查找符合的port
	for _, con := range cons {
		for _, p := range con.Ports {
			if p.ContainerPort == srvPort {
				return srvPort
			}
		}
	}

	return -10000
}

func DeleteEndpoints(batch bool, suffix string) error {
	uri := ""
	getListUri := apiserver.API_server_prefix

	if batch {
		uri = apiserver.API_server_prefix + apiserver.API_delete_endpoints_prefix + suffix
		getListUri += apiserver.API_get_endpoint_by_service_prefix + suffix
	} else {
		uri = apiserver.API_server_prefix + apiserver.API_delete_endpoint_prefix + suffix
		getListUri += apiserver.API_get_endpoint_prefix + suffix
	}

	ep_list := []api_obj.Endpoint{}
	err := network.GetRequestAndParse(getListUri, &ep_list)
	if err != nil {
		fmt.Printf("[ERR/EP Controller/Utils/DeleteEndpoint] Failed to send GET request, %v.\n", err)
		return err
	}
	ep_list_str, err := json.Marshal(ep_list)
	if err != nil {
		fmt.Printf("[ERR/Controller/Utils/DeleteEndpoint] Failed to marshal data, %v.\n", err)
		return err
	}

	_, err = network.DelRequest(uri)
	if err != nil {
		fmt.Printf("[ERR/EP Controller/Utils/DeleteEndpoint] DEL request failed, %v.\n", err)
		return err
	}

	//向proxy发送删除的消息请求。分为批量和非批量。
	for _, ip := range tools.NodesIpMap {
		uri := ip + strconv.Itoa(int(kube_proxy.Port)) + kube_proxy.DeleteEndpoint
		_, err = network.PostRequest(uri, ep_list_str)
		if err != nil {
			fmt.Printf("[ERR/Controller/Utils/DeleteEndpoint] Failed to send POST request, %v.\n", err)
			return err
		}
	}

	return nil
}
