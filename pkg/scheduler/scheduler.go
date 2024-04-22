package scheduler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/message"
	"minik8s/pkg/network"
)

var pollIndex int32 = 0
var lock sync.Mutex

type Scheduler struct {
	consumer   *(message.MsgConsumer)
	producer   *(message.MsgProducer)
	policy     SchedulingPolicy
	apiAddress string
	apiPort    string
}

func (s *Scheduler) MsgHandler(_ *message.MsgDummy) {
	fmt.Printf("[scheduler/MsgHandler] Received message from apiserver!\n")

	podDummy := api_obj.Pod{}

	s.ExecSchedule(podDummy)
}

func (s *Scheduler) ExecSchedule(pod api_obj.Pod) {
	pack, err := s.GetNodes()

	if err != nil {
		fmt.Printf("[ERR/scheduler/ExecSchedule] Failed to get nodes, %s", err)
		return
	}

	avail_pack := []api_obj.Node{}
	//遍历node列表，进行可用性筛选（predicate）
	for _, n := range pack {
		if n.GetCondition() == api_obj.Ready {
			avail_pack = append(avail_pack, n)
		}
	}

	node_chosen := s.DecideNode(pod, avail_pack)
	if node_chosen == "" {
		fmt.Printf("[ERR/scheduler/ExecSchedule] No suitable node to distribute")
		return
	}

	pod.Spec.NodeName = node_chosen
	pod_str, err := json.Marshal(pod)
	if err != nil {
		fmt.Printf("[ERR/scheduler/ExecSchedule] Failed to marshal pod")
		return
	}

	//向apiserver提交更新请求
	uri := s.apiAddress + ":" + s.apiPort + config.API_update_pod
	_, errStr, err := network.PostRequest(uri, pod_str)
	if err != nil {
		fmt.Printf("[ERR/scheduler/ExecSchedule] Failed to update pod to apiserver, %s", err)
		return
	} else if errStr != "" {
		fmt.Printf("[ERR/scheduler/ExecSchedule] Failed to update pod to apiserver, %s", errStr)
		return
	} else {
		fmt.Printf("[scheduler/ExecSchedule] Updated pod to apiserver")
	}

	//向消息队列发送创建pod消息->kubelet
	//TODO:暂定发送topic为kubelet+node名字，且此处的消息为假体
	//实际应为更新后的pod对象。
	msgDummy := message.MsgDummy{}
	s.producer.Produce("kubelet-"+node_chosen, &msgDummy)
}

func (s *Scheduler) DecideNode(pod api_obj.Pod, avail_pack []api_obj.Node) string {
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
	fmt.Printf("[scheduler/SchedulePoll] Node %d chosen for the new pod", i)

	return node_chosen
}

func (s *Scheduler) ScheduleRandom(avail_pack []api_obj.Node) string {
	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(1000) % len(avail_pack)

	node_chosen := avail_pack[i].GetName()
	fmt.Printf("[scheduler/ScheduleRandom] Node %d chosen for the new pod", i)

	return node_chosen
}

func (s *Scheduler) GetNodes() ([]api_obj.Node, error) {
	//向apiServer发送http请求
	uri := s.apiAddress + ":" + s.apiPort + config.API_get_nodes
	var pack []api_obj.Node

	dataStr, errStr, err := network.GetRequest(uri)
	if err != nil {
		fmt.Printf("[ERR/scheduler/GetNodes] GET request failed, %s", err)
		return pack, err
	} else if errStr != "" {
		fmt.Printf("[ERR/scheduler/GetNodes] GET request failed, %s", errStr)
		return pack, err
	}

	err = json.Unmarshal([]byte(dataStr), &pack)
	if err != nil {
		fmt.Printf("[scheduler/GetNodes] Failed to unmarshall data, %s", err)
		return []api_obj.Node{}, err
	}

	return pack, nil
}

func CreateSchedulerInstance() (*Scheduler, error) {
	consumer, _ := message.NewConsumer("scheduler", "default")
	producer := message.NewProducer()

	c := DefaultSchedulerConfig()
	scheduler := &Scheduler{
		consumer:   consumer,
		producer:   producer,
		policy:     c.policy,
		apiAddress: c.apiAddress,
		apiPort:    c.apiPort,
	}

	return scheduler, nil
}

func (s *Scheduler) Run() {
	go s.consumer.Consume([]string{"scheduler"}, s.MsgHandler)
}
