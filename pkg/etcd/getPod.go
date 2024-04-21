package etcd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"my-Minik8s/pkg/api_obj"
	"os"

	"gopkg.in/yaml.v3"
)

func AddPod(filePath string) {
	var etcdWrap *EtcdWrap = nil
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("connect to etcd failed, err:%v\n", err)
	}
	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("connect to etcd failed, err:%v\n", err)
	}
	pod := &api_obj.Pod{}

	err = yaml.Unmarshal(content, pod)
	if err != nil {
		fmt.Printf("connect to etcd failed, err:%v\n", err)
	}

	// 读取的内容转化为json
	jsonBytes, err := json.Marshal(pod)
	if err != nil {
		fmt.Printf("connect to etcd failed, err:%v\n", err)
	}

	// 将 *bytes.Reader 转换为 []byte
	podBytes, err := io.ReadAll(bytes.NewReader(jsonBytes))
	if err != nil {
		fmt.Printf("failed to read pod bytes: %v\n", err)
		return
	}

	etcdWrap.Put(pod.Spec.NodeName, podBytes)
}
