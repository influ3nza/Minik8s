package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/network"
)

func CompareLabels(a map[string]string, b map[string]string) bool {
	for key, val := range a {
		if b[key] != val {
			return false
		}
	}
	return true
}

func CreateEndpoint(srv api_obj.Service, pod api_obj.Pod) error {
	for _, pair := range srv.Spec.Ports {
		port := GetMatchPort(pair.TargetPort, pod.Spec.Containers)
		if port == 0 {
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

		ep_str, err := json.Marshal(ep)
		if err != nil {
			fmt.Printf("[ERR/Controller/Utils/Endpoint] Failed to marshal endpoint, " + err.Error())
			return err
		}

		fmt.Printf("[Controller/Utils/Endpoint] Try to create endpoint %s.\n", ep.MetaData.Name)

		uri := config.API_server_prefix + config.API_add_endpoint
		_, err = network.PostRequest(uri, ep_str)
		if err != nil {
			fmt.Printf("[ERR/Controller/Utils/Endpoint] GET request failed, %v.\n", err)
			return err
		}

		fmt.Printf("[Controller/Utils/Endpoint] Endpoint create request for srv %s:%d, pod %s:%d.\n",
			srv.Spec.ClusterIP, pair.TargetPort, pod.PodStatus.PodIP, port)
	}

	return nil
}

func GetMatchPort(srvPort int32, cons []api_obj.Container) int32 {
	return 10
}

func DeleteEndpoint(batch bool, suffix string) error {
	uri := ""

	if batch {
		uri = config.API_server_prefix + config.API_delete_endpoints_prefix + suffix
	} else {
		uri = config.API_server_prefix + config.API_delete_endpoint_prefix + suffix
	}

	_, err := network.DelRequest(uri)
	if err != nil {
		fmt.Printf("[ERR/EP Controller/Utils/DeleteEndpoint] DEL request failed, %v.\n", err)
		return err
	}

	return nil
}
