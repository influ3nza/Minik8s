package message

import (
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
)

type MsgProducer struct {
	Producer sarama.AsyncProducer
	Sig      chan struct{}
}

func NewProducer() *MsgProducer {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	producer, _ := sarama.NewAsyncProducer([]string{"192.168.1.13:9092"}, config)

	mp := &MsgProducer{
		Producer: producer,
		Sig:      make(chan struct{}),
	}

	go func() {
		for {
			select {
			case success := <-producer.Successes():
				fmt.Printf("[SUCCESS/message/producer] Produced message to topic %s, partition %d, offset %d\n",
					success.Topic, success.Partition, success.Offset)
			case err := <-producer.Errors():
				fmt.Printf("Failed to produce message: %v\n", err)
			case <-mp.Sig:
				return
			}
		}
	}()

	return mp
}

func (mp *MsgProducer) Produce(topic string, msg *Message) {
	msg_str, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("[ERROR/message/producer] Failed to marshal message")
	}

	mp.Producer.Input() <- &sarama.ProducerMessage{Topic: topic, Value: sarama.ByteEncoder(msg_str)}
}

func (mp *MsgProducer) CallScheduleNode(pod_str string) {
	//apiserver -> scheduler
	msg := &Message{
		//TODO: 这里的type是硬编码，需要写进config
		Type:    "ScheduleNode",
		Content: pod_str,
	}

	mp.Produce("scheduler", msg)
}
