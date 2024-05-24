package main

import (
	"fmt"
	"minik8s/pkg/monitor"
)

func Run() error {
	mc := monitor.InitController()
	if mc == nil {
		fmt.Println("Start Failed")
		return fmt.Errorf("gin engine init failed")
	}

	mc.RegisterHandler()

	err := mc.Router.Run(":27500")
	if err != nil {
		return fmt.Errorf("setup Server Failed")
	}
	return nil
}
