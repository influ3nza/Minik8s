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
		fmt.Printf("[ERR/kubectl/parsePod] Failed to open file, err:%v\n", err)
		return err
	}

	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parsePod] Failed to read file, err:%v\n", err)
		return err
	}
	pod := &api_obj.Pod{}

	err = yaml.Unmarshal(content, pod)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parsePod] Failed to unmarshal yaml, err:%v\n", err)
		return err
	}

	pod_str, err := json.Marshal(pod)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parsePod] Failed to marshal pod, err:%v\n", err)
		return err
	}

	//将请求发送给apiserver
	uri := config.API_server_prefix + config.API_add_pod
	_, errStr, err := network.PostRequest(uri, pod_str)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parsePod] Failed to post request, err:%v\n", err)
		return err
	} else if errStr != "" {
		fmt.Printf("[ERR/kubectl/parsePod] Failed to post request, err:%v\n", errStr)
		return nil
	}

	fmt.Printf("[kubectl/parsePod] Send add pod request success!\n")

	return nil
}

func ParseNode(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parseNode] Failed to open file, err:%v\n", err)
		return err
	}

	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parseNode] Failed to read file, err:%v\n", err)
		return err
	}

	node := &api_obj.Node{}
	err = yaml.Unmarshal(content, node)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parseNode] Failed to unmarshal yaml, err:%v\n", err)
		return err
	}

	node_str, err := json.Marshal(node)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parseNode] Failed to marshal pod, err:%v\n", err)
		return err
	}

	//将请求发送给apiserver
	uri := config.API_server_prefix + config.API_add_node
	_, errStr, err := network.PostRequest(uri, node_str)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parseNode] Failed to post request, err:%v\n", err)
		return err
	} else if errStr != "" {
		fmt.Printf("[ERR/kubectl/parseNode] Failed to post request, err:%v\n", errStr)
		return nil
	}

	fmt.Printf("[kubectl/parseNode] Send add node request success!\n")

	return nil
}

func SendObjectTo(jsonStr []byte, kind string) error {
	var suffix string
	switch kind {
	case "pod":
		suffix = config.API_add_pod
	case "node":
		suffix = config.API_add_node
	case "service":
		suffix = config.API_add_service
	}

	uri := config.API_server_prefix + suffix
	_, errStr, err := network.PostRequest(uri, jsonStr)
	if err != nil {
		fmt.Printf("[ERR/kubectl/apply"+kind+"] Failed to send request, err: %s\n", err.Error())
		return err
	} else if errStr != "" {
		fmt.Printf("[ERR/kubectl/apply"+kind+"] Failed to send request, err: %v\n", errStr)
		return nil
	}

	fmt.Printf("[kubectl/apply" + kind + "] Send add " + kind + " request success!\n")

	return nil
}
