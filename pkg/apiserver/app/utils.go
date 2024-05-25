package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/apiserver/controller/utils"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/etcd"
	"minik8s/pkg/kubelet/util"
	"minik8s/pkg/message"
	"minik8s/pkg/network"
	"minik8s/tools"

	"github.com/gin-gonic/gin"
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

	old_pod.MetaData.Annotations = pod.MetaData.Annotations

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
	fmt.Printf("[Apiserver/UpdatePodPhase] Update pod phase: %s -> %s\n", old_pod.PodStatus.Phase, pod.PodStatus.Phase)
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

func AllocateClusterIp(wrap *etcd.EtcdWrap) string {
	clusterIp := "10.1.0." + strconv.Itoa(int(tools.ClusterIpFlag))
	tools.ClusterIpFlag += 1
	//TODO:存入etcd持久化，
	e_key := apiserver.ETCD_service_mark_prefix
	_ = wrap.Put(e_key, []byte(strconv.Itoa(int(tools.ClusterIpFlag))))
	return clusterIp
}

func (s *ApiServer) RefreshNodeIp() error {
	e_key := apiserver.ETCD_node_ip_prefix
	res, err := s.EtcdWrap.GetByPrefix(e_key)
	if err != nil {
		fmt.Printf("[ERR/apiserver/utils/RefreshNodeIp] Failed tp get from etcd, %v", err)
		return err
	}

	for _, kv := range res {
		nodename := kv.Key[strings.LastIndex(kv.Key, "/")+1:]
		fmt.Printf("[Refresh node] Node: %s, %s.\n", nodename, kv.Value)
		tools.NodesIpMap[nodename] = kv.Value
	}

	return nil
}

func (s *ApiServer) GetPodsOfFunction(funcName string) ([]string, error) {
	//搜索所有pod，找出function所在的所有pod的ip地址。
	//所有是function的pod都会在label中有func: function_name的记录。
	pack := []string{}
	e_key := apiserver.ETCD_pod_prefix
	res, err := s.EtcdWrap.GetByPrefix(e_key)
	if err != nil {
		fmt.Printf("[ERR/GetPodsOfFunction] Failed to get from etcd, %v", err)
		return pack, nil
	}

	for _, kv := range res {
		pod := &api_obj.Pod{}
		err := json.Unmarshal([]byte(kv.Value), pod)
		if err != nil {
			fmt.Printf("[ERR/GetPodsOfFunction] Failed to unmarshal data, %v", err)
			return []string{}, nil
		}
		if pod.MetaData.Labels["func"] == funcName && pod.PodStatus.Phase == obj_inner.Running {
			pack = append(pack, pod.PodStatus.PodIP)
		}
	}

	return pack, nil
}

func (s *ApiServer) U_ScaleReplicaSet(funcName string, offset int) error {
	e_key := apiserver.ETCD_replicaset_prefix +
		apiserver.API_default_namespace + "/" + utils.RS_name_prefix + funcName
	fmt.Printf("[U_ScaleRS] e_key: %s\n", e_key)
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		fmt.Printf("[ERR/U_ScaleReplicaSet] Failed to get from etcd, %s.\n", err.Error())
		return err
	}
	if len(res) != 1 {
		fmt.Printf("[ERR/U_ScaleReplicaSet] Found zero or more than one rs.\n")
		return errors.New("found zero or more than one rs")
	}

	rs := &api_obj.ReplicaSet{}
	err = json.Unmarshal([]byte(res[0].Value), rs)
	if err != nil {
		fmt.Printf("[ERR/U_ScaleReplicaSet] Failed to unmarshal data, %s.\n", err.Error())
		return err
	}

	//如果还没创建完毕，则不重复增加。用于冷启动。
	if rs.Status.ReadyReplicas < rs.Spec.Replicas {
		return nil
	}

	rs.Spec.Replicas += offset
	if rs.Spec.Replicas < 0 {
		rs.Spec.Replicas = 0
	}
	if rs.Spec.Replicas > 10 {
		rs.Spec.Replicas = 10
	}

	rs_str, err := json.Marshal(rs)
	if err != nil {
		fmt.Printf("[ERR/U_ScaleReplicaSet] Failed to marshal data, %s.\n", err.Error())
		return err
	}

	err = s.EtcdWrap.Put(e_key, rs_str)
	if err != nil {
		fmt.Printf("[ERR/U_ScaleReplicaSet] Failed to put into etcd, %s.\n", err.Error())
		return err
	}

	return nil
}

func (s *ApiServer) ReadServiceMark() int {
	e_key := apiserver.ETCD_service_mark_prefix
	res, err := s.EtcdWrap.GetByPrefix(e_key)
	if err != nil || len(res) == 0 {
		return 2
	}

	ret, err := strconv.Atoi(res[0].Value)
	if err != nil {
		return 2
	}

	return ret
}

func (s *ApiServer) DynamicCreatePV(pvc *api_obj.PVC) error {
	pv := &api_obj.PV{
		Metadata: obj_inner.ObjectMeta{
			Name:      "pv-" + pvc.Metadata.Name,
			NameSpace: pvc.Metadata.NameSpace,
			Labels:    pvc.Spec.Selector,
		},
		Spec: api_obj.PV_spec{
			Nfs: api_obj.Nfs_bind{
				Path:     "/" + pvc.Metadata.Name,
				ServerIp: util.IpAddressMas,
			},
			AccessMode: api_obj.ReadWriteMany,
		},
	}

	pv_str, err := json.Marshal(pv)
	if err != nil {
		fmt.Printf("[ERR/DynamicCreatePV] Failed to marshal data, %s.\n", err)
		return err
	}

	uri := apiserver.API_server_prefix + apiserver.API_add_pv
	_, err = network.PostRequest(uri, pv_str)
	if err != nil {
		fmt.Printf("[ERR/DynamicCreatePV] Failed to send POST request, %s.\n", err)
		return err
	}

	pvc.Spec.BindPV = pv.Metadata.Name

	return nil
}

// PV:只用于通过PVC改写mount路径。
func (s *ApiServer) RewriteMountPath(pod *api_obj.Pod) error {
	//通过pvc的名字，找到bind的pv的名字，进而确定mount路径。
	//WARN:注意根据nfs，理应在node节点保存一张表格，记录每一个mount dir对应的nfs ip地址，
	//但由于在这里只有一个nfs server，所以不保存表格，采用更简单的设计。

	if len(pod.Spec.Volumes) == 0 {
		return nil
	}

	//这里默认需要绑定pvc的pod只有一个volume
	pvc_name := pod.Spec.Volumes[0].PVCName
	if pvc_name == "" {
		return nil
	}

	e_key := apiserver.ETCD_pvc_prefix + pvc_name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		fmt.Printf("[ERR/RewriteMountPath] Failed to get from etcd, %s\n", err.Error())
		return err
	}
	if len(res) != 1 {
		fmt.Printf("[ERR/RewriteMountPath] Found zero or more than one pvc.\n")
		return errors.New("zero or more than one pvc")
	}

	pvc := &api_obj.PVC{}
	err = json.Unmarshal([]byte(res[0].Value), pvc)
	if err != nil {
		fmt.Printf("[ERR/RewriteMountPath] Failed to unmarshal data, %s\n", err.Error())
		return err
	}

	pv_name := pvc.Spec.BindPV
	if pv_name == "" {
		fmt.Printf("[ERR/RewriteMountPath] PVC must bind to a PV.\n")
		return errors.New("no binding pvc")
	}

	p_key := apiserver.ETCD_pv_prefix + pv_name
	res, err = s.EtcdWrap.Get(p_key)
	if err != nil {
		fmt.Printf("[ERR/RewriteMountPath] Failed to get from etcd, %s\n", err.Error())
		return err
	}
	if len(res) != 1 {
		fmt.Printf("[ERR/RewriteMountPath] Found zero or more than one pv.\n")
		return errors.New("zero or more than one pv")
	}

	pv := &api_obj.PV{}
	err = json.Unmarshal([]byte(res[0].Value), pv)
	if err != nil {
		fmt.Printf("[ERR/RewriteMountPath] Failed to unmarshal data, %s\n", err.Error())
		return err
	}

	pod.Spec.Volumes[0].Path = tools.PV_mount_master_path + pv.Spec.Nfs.Path

	return nil
}

func (s *ApiServer) DeleteRegistry(c *gin.Context) {
	s.EtcdWrap.DeleteByPrefix("/registry")
	s.EtcdWrap.DeleteByPrefix("/pv")
	c.JSON(http.StatusOK, gin.H{
		"data": "Delete all success",
	})
}
