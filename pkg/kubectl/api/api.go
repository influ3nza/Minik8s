package api

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/network"
)

func ParsePod(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Failed to open file, err:%v\n", err)
		return err
	}

	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Failed to read file, err:%v\n", err)
		return err
	}
	pod := &api_obj.Pod{}

	err = yaml.Unmarshal(content, pod)
	if err != nil {
		fmt.Printf("Failed to unmarshal yaml, err:%v\n", err)
		return err
	}

	pod_str, err := json.Marshal(pod)
	if err != nil {
		fmt.Printf("Failed to marshal pod, err:%v\n", err)
		return err
	}

	//将请求发送给apiserver
	uri := config.API_add_pod
	_, errStr, err := network.PostRequest(uri, pod_str)
	if err != nil {
		fmt.Printf("Failed to post request, err:%v\n", err)
		return err
	} else if errStr != "" {
		fmt.Printf("Failed to post request, err:%v\n", errStr)
	}

	fmt.Printf("Send add pod request success!\n")

	return nil
}
