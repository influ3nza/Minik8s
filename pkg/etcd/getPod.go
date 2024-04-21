package etcd

import (
	etcd "go.etcd.io/etcd/client/v3"
	"os"
	"io"
)

func AddPod(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatal(err)
	}
	content, err := io.ReadAll(file)
		if err != nil {
			t.Fatal(err)
		}
	pod := &api_obj.pod{}

	err = yaml.Unmarshal(content, pod)
	if err != nil {
		t.Fatal(err)
	}

	// 读取的内容转化为json
	jsonBytes, err := json.Marshal(pod)

	if err != nil {
		t.Fatal(err)
	}
	podReader := bytes.NewReader(jsonBytes)
	etcd.Put(pod.Spec.NodeName, podReader)
}