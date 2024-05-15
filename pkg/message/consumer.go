package message

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
)

type MsgConsumer struct {
	Consumer sarama.ConsumerGroup
}

func NewConsumer(topic, groupId string) (*MsgConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Return.Errors = true

	// 创建消费者组
	consumer, err := sarama.NewConsumerGroup([]string{"192.168.1.13:9092"}, groupId, config)
	if err != nil {
		return nil, err
	}

	mc := &MsgConsumer{
		Consumer: consumer,
	}

	return mc, nil
}

func (mc *MsgConsumer) Consume(topic []string, callback func(*Message)) {
	handler := &ConsumerHandler{
		callback: callback,
	}

	// 启动消费者组
	go func() {
		for {
			err := mc.Consumer.Consume(context.Background(), topic, handler)
			if err != nil {
				fmt.Printf("Error from consumer: %v\n", err)
			}
		}
	}()
}

type ConsumerHandler struct {
	callback func(*Message)
}

func (h *ConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var dummy Message
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
