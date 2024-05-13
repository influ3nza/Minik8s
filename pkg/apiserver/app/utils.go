package app

import (
	"encoding/json"
	"errors"
	"fmt"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/message"
	"minik8s/pkg/network"
)

func (s *ApiServer) PodNeedRestart(pod api_obj.Pod) {
	//需要重启一个pod，本质上就是将其删除之后再启动一个新的。
	//TODO:这里先不考虑有关于replicaset的问题。其实也可以考虑：如果发现pod隶属于某一个replicaset，
	//则不重启。
	//这里可能是带走PV持久化存储的地方。需要注意。

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
	uri := apiserver.API_server_prefix + apiserver.API_add_pod
	_, err = network.PostRequest(uri, pod_str)
	if err != nil {
		fmt.Printf("[ERR/Apiserver/PodNeedRestart] Failed to POST request, %v.\n", err)
	}

	fmt.Printf("[Apiserver/PodNeedRestart] Create pod request sent.\n")
}

func (s *ApiServer) UpdatePodPhase(pod api_obj.Pod, needCheckRestart bool) (string, error) {
	e_key := apiserver.ETCD_pod_prefix + pod.MetaData.NameSpace + "/" + pod.MetaData.Name
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

	new_pod_str, err := json.Marshal(pod)
	if err != nil {
		fmt.Printf("[ERR/Apiserver/UpdatePodPhase] Failed to marshal data, %v.\n", err)
		return "", err
	}

	if old_pod.PodStatus.PodIP != "" && old_pod.PodStatus.PodIP != pod.PodStatus.PodIP {
		//向ep controller发送update pod的消息。
		ep_msg := &message.Message{
			Type:    message.POD_UPDATE,
			Content: string(new_pod_str),
		}
		s.Producer.Produce(message.TOPIC_EndpointController, ep_msg)
	}
	old_pod.PodStatus.PodIP = pod.PodStatus.PodIP

	fmt.Printf("[Apiserver/UpdatePodPhase] Updated pod ip: %s.\n", pod.PodStatus.PodIP)

	//检查是否重启了
	if needCheckRestart && old_pod.PodStatus.Phase != pod.PodStatus.Phase {
		old_pod.PodStatus.Restarts += 1
	}
	old_pod.PodStatus.Phase = pod.PodStatus.Phase

	//存入etcd中。
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

func (s *ApiServer) UpdateSrvCondition(namespace string, name string) error {
	//在etcd中更新srv的状态。
	e_key := apiserver.ETCD_service_prefix + namespace + "/" + name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		fmt.Printf("[ERR/Apiserver/UpdateSrvCondition] Failed to get from etcd, %v.\n", err)
		return err
	} else if len(res) != 1 {
		fmt.Printf("[ERR/Apiserver/UpdateSrvCondition] Get zero or more than one srv.\n")
		return errors.New("ERR")
	}

	old_srv := &api_obj.Service{}
	err = json.Unmarshal([]byte(res[0].Value), old_srv)
	if err != nil {
		fmt.Printf("[ERR/Apiserver/UpdateSrvCondition] Failed to unmarshal data, %v.\n", err)
		return err
	}

	old_srv.Status.Condition = api_obj.SERVICE_CREATED

	old_str, err := json.Marshal(old_srv)
	if err != nil {
		fmt.Printf("[ERR/Apiserver/UpdateSrvCondition] Failed to marshal data, %v.\n", err)
		return err
	}
	err = s.EtcdWrap.Put(e_key, old_str)
	if err != nil {
		fmt.Printf("[ERR/Apiserver/UpdateSrvCondition] Failed to put into etcd, %v.\n", err)
		return err
	}

	return nil
}
