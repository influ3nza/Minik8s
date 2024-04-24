package etcd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"minik8s/pkg/api_obj"
	"os"

	"gopkg.in/yaml.v3"
)

func AddPod(filePath string, etcdWrap *EtcdWrap) error {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("connect to etcd failed, err:%v\n", err)
		return err
	}
	print("line 20")
	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("connect to etcd failed, err:%v\n", err)
		return err
	}
	print("line26")
	pod := &api_obj.Pod{}
	print()
	print("line29")

	err = yaml.Unmarshal(content, pod)
	if err != nil {
		fmt.Printf("connect to etcd failed, err:%v\n", err)
		return err
	}
	print("line36")

	// 读取的内容转化为json
	jsonBytes, err := json.Marshal(pod)
	if err != nil {
		fmt.Printf("connect to etcd failed, err:%v\n", err)
		return err
	}
	print("line44")

	// 将 *bytes.Reader 转换为 []byte
	podBytes, err := io.ReadAll(bytes.NewReader(jsonBytes))
	if err != nil {
		fmt.Printf("failed to read pod bytes: %v\n", err)
		return err
	}
	fmt.Println("NodeName:", pod.Spec.NodeName)
	fmt.Println("PodBytes:", string(jsonBytes))

	etcdWrap.Put(pod.Spec.NodeName, podBytes)
	return nil
}
