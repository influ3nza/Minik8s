package controller

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/controller/utils"
	"minik8s/pkg/message"
	"minik8s/pkg/network"
)

type ReplicasetController struct {
}

func (rc *ReplicaController) GetAllReplicasetsFromAPIServer() ([]api_obj.ReplicaSet, error) {
	uri := config.API_server_prefix + config.API_get_replicasets
	dataStr, err := network.GetRequest(uri)
	if err != nil {
		fmt.Printf("[ERR/replicasetController/GetAllReplicasetsFromAPIServer] GET request failed, %v.\n", err)
		ec.PrintHandlerWarning()
		return
	}

	var allReplicaSet []api_obj.ReplicaSet
	if dataStr == "" {
		fmt.Printf("[ERR/EndpointController/OnAddService] Not any pod available.\n")
		ec.PrintHandlerWarning()
	} else {
		err = json.Unmarshal([]byte(dataStr), &allReplicaSet)
		if err != nil {
			fmt.Printf("[ERR/ReplicasetController/GetAllReplicasetsFromAPIServer] Failed to unmarshal data, %s.\n", err)
			ec.PrintHandlerWarning()
			return nil,err
		}
		return allReplicaSet, nil
	}
}

func (rc *ReplicaController) routine() {
	//返回所有的pods
	uri := config.API_server_prefix + config.API_get_pods
	dataStr, err := network.GetRequest(uri)
	if err != nil {
		fmt.Printf("[ERR/ReplicaController/routine] GET request failed, %v.\n", err)
		ec.PrintHandlerWarning()
		return
	}
	var allPods []api_obj.Pod
	if dataStr == "" {
		fmt.Printf("[ERR/ReplicaController/routine] Not any pod available.\n")
		ec.PrintHandlerWarning()
	} else {
		err = json.Unmarshal([]byte(dataStr), &allPods)
		if err != nil {
			fmt.Printf("[ERR/ReplicaController/routine] Failed to unmarshal data, %s.\n", err)
			ec.PrintHandlerWarning()
			return
		}
	}

	//返回所有的replicasets
	replicasets, err := rc.GetAllReplicasetsFromAPIServer()
	if err != nil {
		return
	}

	//遍历每一个replicaset
	for _, rs := range replicasets {
		meetRequirementPods := make([]apiObject.Pod, 0)
		//遍历pod，找到存在于replicaset中的pod
		for _, pod := range pods {
			if CheckIfPodMeetRequirement(&pod, rs.Spec.Selector) {
				meetRequirementPods = append(meetRequirementPods, pod)
			}
		}

		//根据replicaset要求的数量，删减replicaset中的pod
		//如果小了就增加
		if len(meetRequirementPods) < rs.Spec.Replicas {
			rc.AddReplicaPodsNums(&rs.Metadata, &rs.Spec.Template, rs.Spec.Replicas - len(meetRequirementPods))
		}
		//如果大了就减小 
		else if len(meetRequirementPods) > rs.Spec.Replicas {
			rc.ReduceReplicaPodsNums(meetRequirementPods, len(meetRequirementPods) - rs.Spec.Replicas)
		}

		// 3. 根据选择好的pod的状态，更新replicasets的状态
		// 注意，以上对replicaset的修改不会马上反映在replicaset的status里
		rc.UpdateReplicaSet(meetRequirementPods, &rs)
	}
}

// 增加pod的数量
func (rc *replicaController) AddReplicaPodsNums(replicaset *obj_inner.ObjectMeta, pod *api_obj.PodTemplate, num int) error {
	uri := config.API_server_prefix + config.API_add_pod
	podNew := api_obj.Pod{}
	podNew.MetaData = pod.MetaData
	podNew.ApiVersion = "v1"
	podNew.Kind = "Pod"
	podNew.Spec = pod.Spec
	podNew.MetaData.Labels["replicaset_name"] = replicaset.Name
	podNew.MetaData.Labels["replicaset_namespace"] = replicaset.NameSpace
	podNew.Message.Labels["replicaset_uuid"] = replicaset.UUID

	podName := newPod.MetaData.Name

	for i := 0; i < num; i++ {
		rand.Seed(time.Now().UnixNano())
		randomNumber := rand.Intn(1000)
		randomString := strconv.Itoa(randomNumber)
		podNew.MetaData.Name = podName + '-' + randomString + "rsCreate" + num
		
		podJson, err := json.Marshal(podNew)
		if err != nil {
			fmt.Printf("[ERR/replicasetController/Addpod] Failed to marshal pod, %v.\n", err)
			return
		}

		_, err = network.PostRequest(uri, podJson)
		if err != nil {
			fmt.Printf("[ERR/replicasetController/AddReplicaPodsNums] Failed to post request, err:%v\n", err)
			return err
		}

	}
	fmt.Printf("[replicasetController/AddReplicaPodsNums] Send add pod request success!\n")

	return nil

}

func (rc *replicaController) ReduceReplicaPodsNums(pods []api_obj.Pod, num int) error {
	uri := config.API_server_prefix + config.API_delete_pod
	for i := 0; i < num; i++ {
		podJson, err := json.Marshal(pods[i])
		if err != nil {
			fmt.Printf("[ERR/replicasetController/ReduceReplicaPodsNums] Failed to marshal pod, %v.\n", err)
			return
		}

		_, err = network.PostRequest(uri, podJson)
		if err != nil {
			fmt.Printf("[ERR/replicasetController/ReduceReplicaPodsNums] Failed to post request, err:%v\n", err)
			return err
		}
	}
	fmt.Printf("[replicasetController/ReduceReplicaPodsNums] Send delete pod request success!\n")
	return nil

}

func (rc *replicaController) UpdateReplicaSet(pods []api_obj.Pod, rs *api_obj.Replicaset) error {
	uri := config.API_server_prefix + config.API_update_replicaset
	newReplicaSet := api_obj.ReplicaSet{}

	ready := 0
	for _, pod := range pods {
		if pod.PodStatus.Phase == obj_inner.Running || obj_inner.Succeeded {
			ready += 1
		}
	}
	newReplicaSet.Status.Replicas = rs.Spec.Replicas
	newReplicaSet.Status.ReadyReplicas = ready

	replicasetSetJson, err := json.Marshal(newReplicaSet)
	if err != nil {
		fmt.Printf("[ERR/replicasetController/UpdateReplicaSet] Failed to marshal pod, %v.\n", err)
		return
	}
	_, err = network.PostRequest(uri, replicasetSetJson)
	if err != nil {
		fmt.Printf("[ERR/replicasetController/UpdateReplicaSet] Failed to post request, err:%v\n", err)
		return err
	}

}


func (rc *replicaController) CheckIfPodMeetRequirement(pod *api_obj.Pod, selectors map[string]string) bool {
	// 只要pod的label中有一个key-value对与selector中的key-value对相同，就认为pod满足要求
	podLabel := pod.Metadata.Labels
	for key, value := range selectors {
		if podLabel[key] != value {
			return false
		} else {
			continue
		}
	}
	return true
}