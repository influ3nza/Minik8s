package etcd

import (
	"testing"
	"my-Minik8s/pkg/etcd"
)

func TestAddPod(t *testing.T) {
	filePath := "./testfile/Pod-1.yaml"
	err := AddPod(filePath)
	if err != nil {
		t.Errorf("AddPod returned an error: %v", err)
	}
}
