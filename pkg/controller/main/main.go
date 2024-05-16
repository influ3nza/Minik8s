package main

import (
	"fmt"
	"minik8s/pkg/controller"
	"time"
)

func main() {
	ep, err := controller.CreateEndpointControllerInstance()
	if err != nil {
		fmt.Printf("[Controller/MAIN] Failed to create ep controller.")
	}

	go ep.Run()

	for {
		time.Sleep(100 * time.Millisecond)
	}
}
