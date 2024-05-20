package server

import (
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/network"
)

func (s *SL_server) GetPodsFromApiserver() ([]api_obj.Pod, error) {
	uri := apiserver.API_server_prefix + apiserver.API_get_pods

	var pack []api_obj.Pod
	err := network.GetRequestAndParse(uri, &pack)
	if err != nil {
		fmt.Printf("[ERR/Serverless/utils] Failed to send GET request, %v.\n", err)
		return nil, err
	}

	return pack, nil
}

func (s *SL_server) GeneratePodMap(pods []api_obj.Pod) {
	
}
