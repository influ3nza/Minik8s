package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"minik8s/pkg/kubelet/application"
	"minik8s/pkg/kubelet/util"
	"os"
)

func GetConfigFromFile(file string) (*application.Kubelet, error) {
	File, err := os.Open(file)
	if err != nil {
		fmt.Printf("[ERR/kubelet] Failed to open file, err:%v\n", err)
		return nil, err
	}

	content, err := io.ReadAll(File)
	if err != nil {
		fmt.Printf("[ERR/kubelet] Failed to read file, err:%v\n", err)
		return nil, err
	}

	config := &util.KubeConfig{}
	err = yaml.Unmarshal(content, config)
	if err != nil {
		fmt.Printf("[ERR/kubelet] Failed to unmarshal yaml, err:%v\n", err)
		return nil, err
	}

	return application.InitKubelet(*config), nil
}

func main() {
	if len(os.Args) > 2 {
		return
	}
	var server *application.Kubelet = nil
	if len(os.Args) == 1 {
		server = application.InitKubeletDefault()
	} else {
		var err error
		server, err = GetConfigFromFile(os.Args[1])
		if err != nil {
			fmt.Println("Init Kubelet Failed ", err.Error())
			return
		}
	}

	server.Run()
}
