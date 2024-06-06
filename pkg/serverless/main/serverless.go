package main

import (
	"fmt"
	"minik8s/pkg/serverless/server"
)

func main() {
	server, err := server.CreateNewSLServerInstance()
	if err != nil {
		fmt.Printf("[ERR/Serverless/Server] Error creating server.\n")
	}

	server.Run()
}
