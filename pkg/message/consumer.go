package message

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
)

type MsgConsumer struct {
	consumer sarama.ConsumerGroup
	wg       *sync.WaitGroup
}

func NewConsumer(topic, groupId string) (*MsgConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Return.Errors = true

	// 创建消费者组
	consumer, err := sarama.NewConsumerGroup([]string{"localhost:9092"}, groupId, config)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range consumer.Errors() {
			fmt.Printf("Consumer error: %v\n", err)
		}
	}()

	return &MsgConsumer{
		consumer: consumer,
		wg:       &wg,
	}, nil
}

func (mc *MsgConsumer) Consume(topic []string, callback func(*MsgDummy)) {
	handler := &ConsumerHandler{
		callback: callback,
	}

	// 启动消费者组
	go func() {
		for {
			err := mc.consumer.Consume(context.Background(), topic, handler)
			if err != nil {
				fmt.Printf("Error from consumer: %v\n", err)
			}
		}
	}()
}

type ConsumerHandler struct {
	callback func(*MsgDummy)
}

func (h *ConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var dummy MsgDummy
		err := json.Unmarshal(msg.Value, &dummy)
		if err != nil {
			fmt.Printf("Error unmarshalling message: %v\n", err)
			continue
		}
		h.callback(&dummy)
		session.MarkMessage(msg, "")
	}
	return nil
}
