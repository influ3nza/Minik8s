package message

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type MsgConsumer struct {
	consumer *kafka.Reader
}

func NewConsumer(topic, groupId string) *MsgConsumer {
	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{"localhost:9092"},
		Topic:       topic,
		GroupID:     groupId,
		StartOffset: kafka.FirstOffset,
	})

	return &MsgConsumer{
		consumer: consumer,
	}
}

func (mc *MsgConsumer) consume(callback func(*MsgDummy)) {

	for {
		msg, err := mc.consumer.ReadMessage(context.Background())
		if err != nil {
			fmt.Printf("[ERROR/message/consumer] Failed to read message: %s\n", err)
		}

		dummy := new(MsgDummy)
		err = json.Unmarshal(msg.Value, dummy)
		if err != nil {
			fmt.Printf("[ERROR/message/consumer] Failed to unmarshal message: %s\n", err)
			return
		}

		callback(dummy)
	}
}
