package scheduler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/message"
	"minik8s/pkg/network"
	"minik8s/pkg/config/apiserver"
)

var pollIndex int32 = 0
var lock sync.Mutex

type Scheduler struct {
	Consumer   *(message.MsgConsumer)
	Producer   *(message.MsgProducer)
	policy     SchedulingPolicy
	apiAddress string
	apiPort    string
}

func (s *Scheduler) MsgHandler(msg *message.Message) {
	fmt.Printf("[scheduler/MsgHandler] Received message from apiserver!\n")

	pod_str := msg.Content
	new_pod := &api_obj.Pod{}

	err := json.Unmarshal([]byte(pod_str), new_pod)
	if err != nil {
		fmt.Printf("[ERR/scheduler/MsgHandler] Failed to marshal pod, " + err.Error())
		return
	}

	s.ExecSchedule(new_pod)
}

func (s *Scheduler) ExecSchedule(pod *api_obj.Pod) {
	pack, err := s.GetNodes()

	if err != nil {
		fmt.Printf("[ERR/scheduler/ExecSchedule] Failed to get nodes, %s\n", err)
		return
	}

	avail_pack := []api_obj.Node{}
	//遍历node列表，进行可用性筛选（predicate）
	for _, n := range pack {
		if n.GetCondition() == api_obj.Ready {
			avail_pack = append(avail_pack, n)
		}
	}

	fmt.Printf("[scheduler/ExecSchedule] Available node count: %d\n", len(avail_pack))

	//挑选唯一的node
	node_chosen := s.DecideNode(pod, avail_pack)
	if node_chosen == "" {
		fmt.Printf("[ERR/scheduler/ExecSchedule] No suitable node to distribute.\n")
		return
	}

	pod.Spec.NodeName = node_chosen
	pod_str, err := json.Marshal(pod)
	if err != nil {
		fmt.Printf("[ERR/scheduler/ExecSchedule] Failed to marshal pod.\n")
		return
	}

	//向apiserver提交更新请求
	uri := apiserver.API_server_prefix + apiserver.API_update_pod
	dataStr, err := network.PostRequest(uri, pod_str)
	if err != nil {
		fmt.Printf("[ERR/scheduler/ExecSchedule] Failed to update pod to apiserver, %s.\n", err)
		return
	} else {
		//TODO: 合并后需要修改这里。
		uri = dataStr + "/pod/AddPod"
		_, err = network.PostRequest(uri, pod_str)
		if err != nil {
			fmt.Printf("[ERR/scheduler/ExecSchedule] Failed to send request to node, %s.\n", err)
			return
		}
	}
}

func (s *Scheduler) DecideNode(pod *api_obj.Pod, avail_pack []api_obj.Node) string {
	//指定席
	requested := pod.Spec.NodeName
	if requested != "" {
		for _, n := range avail_pack {
			if n.GetName() == requested {
				return requested
			}
		}
	}

	//调度策略
	switch s.policy {
	case Poll:
		return s.SchedulePoll(avail_pack)
	case Random:
		return s.ScheduleRandom(avail_pack)
	default:
		return ""
	}
}

func (s *Scheduler) SchedulePoll(avail_pack []api_obj.Node) string {
	lock.Lock()
	defer lock.Unlock()

	node_cnt := len(avail_pack)
	if node_cnt == 0 {
		return ""
	}

	i := pollIndex % int32(node_cnt)
	node_chosen := avail_pack[i].GetName()
	pollIndex += 1
	fmt.Printf("[scheduler/SchedulePoll] Node [%d]:%s chosen for the new pod.\n", i, node_chosen)

	return node_chosen
}

func (s *Scheduler) ScheduleRandom(avail_pack []api_obj.Node) string {
	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(1000) % len(avail_pack)

	node_chosen := avail_pack[i].GetName()
	fmt.Printf("[scheduler/ScheduleRandom] Node [%d]:%s chosen for the new pod.\n", i, node_chosen)

	return node_chosen
}

func (s *Scheduler) GetNodes() ([]api_obj.Node, error) {
	//向apiServer发送http请求
	uri := apiserver.API_server_prefix + apiserver.API_get_nodes
	var pack []api_obj.Node

	dataStr, err := network.GetRequest(uri)
	if err != nil {
		fmt.Printf("[ERR/scheduler/GetNodes] GET request failed, %s.\n", err)
		return pack, err
	}

	if dataStr == "" {
		fmt.Printf("[ERR/scheduler/GetNodes] Not any node available.\n")
		return []api_obj.Node{}, nil
	}

	fmt.Printf("[scheduler/GetNodes] Sent request to apiserver!\n")

	err = json.Unmarshal([]byte(dataStr), &pack)
	if err != nil {
		fmt.Printf("[ERR/scheduler/GetNodes] Failed to unmarshall data, %s.\n", err)
		return []api_obj.Node{}, err
	}

	return pack, nil
}

func CreateSchedulerInstance() (*Scheduler, error) {
	consumer, err := message.NewConsumer("scheduler", "scheduler")
	producer := message.NewProducer()

	c := DefaultSchedulerConfig()
	scheduler := &Scheduler{
		Consumer:   consumer,
		Producer:   producer,
		policy:     c.policy,
		apiAddress: c.apiAddress,
		apiPort:    c.apiPort,
	}

	return scheduler, err
}

func (s *Scheduler) Run() {
	go s.Consumer.Consume([]string{"scheduler"}, s.MsgHandler)
}
