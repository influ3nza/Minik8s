package scheduler

import (
	"fmt"

	"minik8s/pkg/message"
)

type Scheduler struct {
	consumer   *(message.MsgConsumer)
	producer   *(message.MsgProducer)
	policy     SchedulingPolicy
	apiAddress string
	apiPort    string
}

func (s *Scheduler) MsgHandler(*message.MsgDummy) {
	fmt.Printf("[Scheduler] Received message from apiserver!\n")
}

func (s *Scheduler) GetNodes() {
	//向apiServer发送http请求

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
