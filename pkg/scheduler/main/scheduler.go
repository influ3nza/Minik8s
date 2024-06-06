package main

import (
	"fmt"
	"minik8s/pkg/scheduler"
	"time"
)

func main() {
	sche, err := scheduler.CreateSchedulerInstance()
	if err != nil {
		fmt.Printf("[Scheduler/MAIN] Failed to create scheduler.")
	}

	go sche.Run()

	for {
		time.Sleep(100 * time.Millisecond)
	}
}
