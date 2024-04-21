package message

import (
	"fmt"
	"strconv"
	"testing"
)

var producerDummy *MsgProducer = nil
var consumerDummy *MsgConsumer = nil
var cnt = 0

func sampleHandler(msg *MsgDummy) {
	fmt.Printf("Received %s, hello.\n", msg.Val)
	cnt++
	if cnt == 10 {
		consumerDummy.consumer.Close()
	}
}

func TestMain(m *testing.M) {
	producerDummy = NewProducer("testMsg")
	consumerDummy = NewConsumer("testMsg", "default")
	go consumerDummy.Consume(sampleHandler)
	m.Run()
}

func TestSend(t *testing.T) {
	for i := 0; i < 10; i++ {
		msg := &MsgDummy{
			Type: "default",
			Key:  "default",
			Val:  "world" + strconv.Itoa(i),
		}

		producerDummy.Produce(msg)
		fmt.Printf("Send message %d\n", i)
	}

	producerDummy.producer.Close()
}
