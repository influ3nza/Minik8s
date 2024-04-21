package scheduler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/apiserver/config"

	"minik8s/pkg/message"
)

type Scheduler struct {
	consumer   *(message.MsgConsumer)
	producer   *(message.MsgProducer)
	policy     SchedulingPolicy
	apiAddress string
	apiPort    string
}

func (s *Scheduler) MsgHandler(_ *message.MsgDummy) {
	fmt.Printf("[Scheduler] Received message from apiserver!\n")
	s.ExecSchedule()
}

func (s *Scheduler) ExecSchedule() {
	pack, err := s.GetNodes()

	for _, kv := range pack {
		fmt.Printf("Get node %s, %s\n", kv.UUID, kv.Val)
	}

	if err != nil {
		_ = fmt.Errorf("[SCHEDULER/execSchedule] Fail to get all nodes")
	} else {
		fmt.Printf("[SCHEDULER/execSchedule] Reached here.\n")
	}
	//TODO:本函数传入pod对象，接下来筛选存活node以及分配最合适的node
}

func (s *Scheduler) GetNodes() ([]api_obj.NodeDummy, error) {
	//向apiServer发送http请求
	uri := s.apiAddress + ":" + s.apiPort + config.API_get_nodes

	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	_ = decoder.Decode(&result)
	data := result["data"]
	dataStr := fmt.Sprint(data)

	var pack []api_obj.NodeDummy
	_ = json.Unmarshal([]byte(dataStr), &pack)

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
