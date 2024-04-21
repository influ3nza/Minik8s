package message

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

var producerDummy *MsgProducer = nil
var consumerDummy *MsgConsumer = nil

func SampleHandler(msg *MsgDummy) {
	fmt.Printf("Received %s, hello.\n", msg.Val)
}

func TestMain(m *testing.M) {
	producerDummy = NewProducer()
	consumerDummy, _ = NewConsumer("testMsg", "default")
	consumerDummy.Consume([]string{"testMsg"}, SampleHandler)
	m.Run()
}

func TestSend(t *testing.T) {
	for i := 0; i < 10; i++ {
		msg := &MsgDummy{
			Type: "default",
			Key:  "default",
			Val:  "world" + strconv.Itoa(i),
		}
		producerDummy.Produce("testMsg", msg)
		fmt.Printf("Send message %d\n", i)
	}

	for {
		time.Sleep(1 * time.Second)
	}
}
