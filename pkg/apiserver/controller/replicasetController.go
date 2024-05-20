package controller

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/network"
)

type ReplicasetController struct {
}

var (
	timedelay    = 10 * time.Second
	timeinterval = []time.Duration{10 * time.Second}
)

func (rc *ReplicasetController) PrintHandlerWarning() {
	fmt.Printf("[WARN/ReplicasetController] Error in message handler, the system may not be working properly!\n")
}

func CreatereplicasetControllerInstance() (*ReplicasetController, error) {
	return &ReplicasetController{}, nil
}

func (rc *ReplicasetController) Run() {
	rc.execute(timedelay, timeinterval, rc.watch)
}

func (rc *ReplicasetController) execute(delay time.Duration, interval []time.Duration, callback callback) {
	if len(interval) == 0 {
		return
	}
	<-time.After(delay)
	for {
		for _, inter := range interval {
			callback()
			<-time.After(inter)
		}
	}
}

func (rc *ReplicasetController) CheckPod(pod *api_obj.Pod, selectors map[string]string) bool {
	podLabel := pod.MetaData.Labels
	//may have bug,我的思路是只要有一个label匹配key-value就返回true
	for key, value := range selectors {
		if podLabel[key] == value {
			return true
		}
	}
	return false
}

func (rc *ReplicasetController) GetAllReplicasets() ([]api_obj.ReplicaSet, error) {
	uri := apiserver.API_server_prefix + apiserver.API_get_replicasets
	dataStr, err := network.GetRequest(uri)
	if err != nil {
		fmt.Printf("[ERR/ReplicasetController/GetAllReplicasets] GET request failed, %v.\n", err)
		rc.PrintHandlerWarning()
		return nil, err
	}

	var rss []api_obj.ReplicaSet
	if dataStr == "" {
		fmt.Printf("[ERR/ReplicasetController/GETALL] Not any Replicaset available.\n")
		rc.PrintHandlerWarning()
		return nil, err
	} else {
		err = json.Unmarshal([]byte(dataStr), &rss)
		if err != nil {
			fmt.Printf("[ERR/ReplicasetController/GetAllReplicasets] Failed to unmarshal data, %s.\n", err)
			rc.PrintHandlerWarning()
			return nil, err
		}
		return rss, nil
	}
}

func (rc *ReplicasetController) watch() {
	//返回所有的pods
	uri := apiserver.API_server_prefix + apiserver.API_get_pods
	dataStr, err := network.GetRequest(uri)
	if err != nil {
		fmt.Printf("[ERR/ReplicaController/watch] GET request failed, %v.\n", err)
		rc.PrintHandlerWarning()
		return
	}
	var allPods []api_obj.Pod
	if dataStr == "" {
		fmt.Printf("[ERR/ReplicaController/watch] Not any pod available.\n")
		rc.PrintHandlerWarning()
	} else {
		err = json.Unmarshal([]byte(dataStr), &allPods)
		if err != nil {
			fmt.Printf("[ERR/ReplicaController/watch] Failed to unmarshal data, %s.\n", err)
			rc.PrintHandlerWarning()
			return
		}
	}

	//返回所有的replicasets
	replicasets, err := rc.GetAllReplicasets()
	if err != nil {
		return
	}

	//遍历每一个replicaset
	for _, rs := range replicasets {
		correspondPods := make([]api_obj.Pod, 0)
		//遍历pod，找到存在于replicaset中的pod
		for _, pod := range allPods {
			if CheckPod(&pod, rs.Spec.Selector) {
				correspondPods = append(correspondPods, pod)
			}
		}

		//根据replicaset要求的数量，删减replicaset中的pod
		//如果小了就增加
		if len(correspondPods) < rs.Spec.Replicas {
			rc.AddReplicaPods(&rs.MetaData, &rs.Spec.Template, rs.Spec.Replicas-len(correspondPods))
		} else if len(correspondPods) > rs.Spec.Replicas {
			rc.ReduceReplicaPods(correspondPods, len(correspondPods)-rs.Spec.Replicas)
		}

		// 3. 根据选择好的pod的状态，更新replicasets的状态
		// 注意，以上对replicaset的修改不会马上反映在replicaset的status里
		rc.UpdateReplicaSet(correspondPods, &rs)
	}
}

// 增加pod的数量
func (rc *ReplicasetController) AddReplicaPods(replicaset *obj_inner.ObjectMeta, pod *api_obj.PodTemplate, num int) error {
	uri := apiserver.API_server_prefix + apiserver.API_add_pod
	podNew := api_obj.Pod{}
	podNew.MetaData = pod.Metadata
	podNew.ApiVersion = "v1"
	podNew.Kind = "Pod"
	podNew.Spec = pod.Spec
	podNew.MetaData.Labels["replicaset_name"] = replicaset.Name
	podNew.MetaData.Labels["replicaset_namespace"] = replicaset.NameSpace
	podNew.MetaData.Labels["replicaset_uuid"] = replicaset.UUID

	podName := podNew.MetaData.Name

	for i := 0; i < num; i++ {
		rand.Seed(time.Now().UnixNano())
		randomNumber := rand.Intn(1000)
		randomString := strconv.Itoa(randomNumber)
		podNew.MetaData.Name = podName + "-" + randomString + "rsCreate" + strconv.Itoa(num)

		podJson, err := json.Marshal(podNew)
		if err != nil {
			fmt.Printf("[ERR/replicasetController/AddReplicaPods] Failed to marshal pod, %v.\n", err)
			return err
		}

		_, err = network.PostRequest(uri, podJson)
		if err != nil {
			fmt.Printf("[ERR/replicasetController/AddReplicaPods] Failed to post request, err:%v\n", err)
			return err
		}

	}
	fmt.Printf("[replicasetController/AddReplicaPods] Send add pod request success!\n")

	return nil

}

func (rc *ReplicasetController) ReduceReplicaPods(pods []api_obj.Pod, num int) error {
	for i := 0; i < num; i++ {
		namespace := pods[i].MetaData.NameSpace
		name := pods[i].MetaData.Name
		uri := apiserver.API_server_prefix + apiserver.API_delete_pod_prefix + namespace + "/" + name
		_, err := json.Marshal(pods[i])
		if err != nil {
			fmt.Printf("[ERR/replicasetController/ReduceReplicaPods] Failed to marshal pod, %v.\n", err)
			return err
		}

		_, err = network.DelRequest(uri)
		if err != nil {
			fmt.Printf("[ERR/replicasetController/ReduceReplicaPods] Failed to post request, err:%v\n", err)
			return err
		}
	}
	fmt.Printf("[replicasetController/ReduceReplicaPods] Send delete pod request success!\n")
	return nil
}

// replicaset通知apiserver去更新，对应的函数是apiserver如何更新
func (rc *ReplicasetController) UpdateReplicaSet(pods []api_obj.Pod, rs *api_obj.ReplicaSet) error {
	uri := apiserver.API_server_prefix + apiserver.API_update_replicaset
	newReplicaSet := api_obj.ReplicaSet{}

	ready := 0
	for _, pod := range pods {
		if pod.PodStatus.Phase == obj_inner.Running || pod.PodStatus.Phase == obj_inner.Succeeded {
			ready += 1
		}
	}
	newReplicaSet = *rs
	newReplicaSet.Status.Replicas = rs.Spec.Replicas
	newReplicaSet.Status.ReadyReplicas = ready

	replicasetStr, err := json.Marshal(newReplicaSet)
	if err != nil {
		fmt.Printf("[ERR/replicasetController/UpdateReplicaSet] Failed to marshal replicaset, %v.\n", err)
		return err
	}
	_, err = network.PostRequest(uri, replicasetStr)
	if err != nil {
		fmt.Printf("[ERR/replicasetController/UpdateReplicaSet] Failed to post request, err:%v\n", err)
		return err
	}
	return nil
}

//todo:发送创建create请求
