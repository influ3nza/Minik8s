package etcd

import (
	"testing"
	"time"
)

func TestAddPod(t *testing.T) {
	filePath := "./testfile/Pod-1.yaml"
	print("test hello")
	etcdWrap, err := CreateEtcdInstance([]string{"http://localhost:50000"}, 5*time.Second)

	if err != nil {
		t.Errorf("AddPod returned an error: %v", err)
	}
	err = AddPod(filePath, etcdWrap)
	print("success addpod")
	if err != nil {
		t.Errorf("AddPod returned an error: %v", err)
	}
}
