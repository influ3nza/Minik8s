package message

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type MsgProducer struct {
	producer *kafka.Writer
}

func NewProducer(topic string) *MsgProducer {
	producer := &kafka.Writer{
		Addr:                   kafka.TCP("localhost:9092"), //不定长参数，支持传入多个broker的ip:port
		Topic:                  topic,                       //为所有message指定统一的topic。如果这里不指定统一的Topic，则创建kafka.Message{}时需要分别指定Topic
		Balancer:               &kafka.Hash{},               //把message的key进行hash，确定partition
		WriteTimeout:           1 * time.Second,             //设定写超时
		RequiredAcks:           kafka.RequireNone,           //RequireNone不需要等待ack返回，效率最高，安全性最低；RequireOne只需要确保Leader写入成功就可以发送下一条消息；RequiredAcks需要确保Leader和所有Follower都写入成功才可以发送下一条消息。
		AllowAutoTopicCreation: true,                        //Topic不存在时自动创建。生产环境中一般设为false，由运维管理员创建Topic并配置partition数目
	}

	return &MsgProducer{
		producer: producer,
	}
}

func (mp *MsgProducer) Produce(msg *MsgDummy) {
	dummy, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("[ERROR/message/producer] Failed to marshal message")
	}

	err = mp.producer.WriteMessages(context.Background(), kafka.Message{Value: dummy})
	if err != nil {
		fmt.Printf("[ERROR/message/producer] Failed to write message, %s\n", err)

		// https://stackoverflow.com/questions/35788697/leader-not-available-kafka-in-console-producer
		// 首次创建的topic发送消息时，第一次一定会失败并显示以上报错。
		_ = mp.producer.WriteMessages(context.Background(), kafka.Message{Value: dummy})
	}
}
