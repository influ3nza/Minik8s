package app

import (
	"encoding/json"
	"errors"
	"fmt"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/network"
)

func (s *ApiServer) PodNeedRestart(pod api_obj.Pod) {
	//需要重启一个pod，本质上就是将其删除之后再启动一个新的。
	//TODO:这里先不考虑有关于replicaset的问题。其实也可以考虑：如果发现pod隶属于某一个replicaset，
	//则不重启。
	//TODO: 这里可能是带走PV持久化存储的地方。需要注意。

	//对pod的某些信息进行修改。
	pod.PodStatus = api_obj.PodStatus{}
	pod.PodStatus.Phase = obj_inner.Pending
	pod.Spec.NodeName = ""

	pod_str, err := json.Marshal(pod)
	if err != nil {
		fmt.Printf("[ERR/Apiserver/PodNeedRestart] Failed to marshal data, %v.\n", err)
		return
	}

	//再次模拟一个pod创建请求。
	uri := config.API_server_prefix + config.API_add_pod
	_, err = network.PostRequest(uri, pod_str)
	if err != nil {
		fmt.Printf("[ERR/Apiserver/PodNeedRestart] Failed to POST request, %v.\n", err)
	}

	fmt.Printf("[Apiserver/PodNeedRestart] Create pod request sent.\n")
}

func (s *ApiServer) UpdatePodPhase(pod api_obj.Pod) (string, error) {
	e_key := config.ETCD_pod_prefix + pod.MetaData.NameSpace + "/" + pod.MetaData.Name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		fmt.Printf("[ERR/Apiserver/UpdatePodPhase] Failed to get from etcd, %v.\n", err)
		return "", err
	} else if len(res) != 1 {
		fmt.Printf("[ERR/Apiserver/UpdatePodPhase] Get zero or more than one pod.\n")
		return "", errors.New("ERR")
	}

	old_pod := &api_obj.Pod{}
	err = json.Unmarshal([]byte(res[0].Value), old_pod)
	if err != nil {
		fmt.Printf("[ERR/Apiserver/UpdatePodPhase] Failed to unmarshal data, %v.\n", err)
		return "", err
	}

	old_pod.PodStatus.PodIP = pod.PodStatus.PodIP
	old_pod.PodStatus.Phase = pod.PodStatus.Phase
	old_str, err := json.Marshal(old_pod)
	if err != nil {
		fmt.Printf("[ERR/Apiserver/UpdatePodPhase] Failed to marshal data, %v.\n", err)
		return "", err
	}

	err = s.EtcdWrap.Put(e_key, old_str)
	if err != nil {
		fmt.Printf("[ERR/Apiserver/UpdatePodPhase] Failed to put into etcd, %v.\n", err)
		return "", err
	}

	return string(old_str), nil
}
