package main

import (
	"fmt"
	"minik8s/pkg/apiserver/app"
	"minik8s/pkg/config/apiserver"
)

func main() {
	server, err := app.CreateApiServerInstance(apiserver.DefaultServerConfig())
	if err != nil {
		fmt.Printf("[Apiserver/MAIN] Failed to create apiserver.")
	}

	_ = server.Run()
}
