package scheduler

import (
	"encoding/json"
	"fmt"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/message"
	"minik8s/pkg/network"
)

var pollIndex int32 = 0

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
	//向apiserver提交更新请求
	//向消息队列发送创建pod消息->kubelet
}

func (s *Scheduler) DecideNode(pod api_obj.Pod, avail_pack []api_obj.Node) string {
	//指定，数量为0，返回""
}

func (s *Scheduler) GetNodes() ([]api_obj.Node, error) {
	//向apiServer发送http请求
	uri := s.apiAddress + ":" + s.apiPort + config.API_get_nodes
	var pack []api_obj.Node

	dataStr, err := network.GetRequest(uri)
	if err != nil {
		fmt.Printf("[ERR/scheduler/GetNodes] GET request failed, %s", err)
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
