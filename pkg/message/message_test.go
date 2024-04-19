package message

import (
	"fmt"
	"strconv"
	"testing"
)

var producerDummy *MsgProducer = nil
var consumerDummy *MsgConsumer = nil

func sampleHandler(msg *MsgDummy) {
	fmt.Printf("Received %s, hello.\n", msg.Val)
}

func TestMain(m *testing.M) {
	producerDummy = NewProducer("testMsg")
	consumerDummy = NewConsumer("testMsg", "default")
	go consumerDummy.consume(sampleHandler)
	m.Run()
}

func TestSend(t *testing.T) {
	for i := 0; i < 10; i++ {
		msg := &MsgDummy{
			Type: "default",
			Key:  "default",
			Val:  "world" + strconv.Itoa(i),
		}

		producerDummy.produce(msg)
	}

	producerDummy.producer.Close()
	consumerDummy.consumer.Close()
}
