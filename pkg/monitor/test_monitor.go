package monitor

import (
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"os"
	"testing"
	"time"
)

func TestGenerateNode(t *testing.T) {
	hostName, _ := os.Hostname()
	node := &api_obj.Node{
		APIVersion: "v1",
		NodeMetadata: obj_inner.ObjectMeta{
			Name: hostName,
			Labels: map[string]string{
				"name": hostName,
			},
		},
		Kind: "node",
		NodeStatus: api_obj.NodeStatus{
			Addresses: api_obj.Address{
				HostName:   hostName,
				ExternalIp: "10.119.13.178",
				InternalIp: "192.168.1.13",
			},
			Condition: api_obj.Ready,
			Capacity: map[string]string{
				"cpu":    "2",
				"memory": "4GiB",
			},
			Allocatable: map[string]string{},
			UpdateTime:  time.Now(),
		},
	}

	nodeStruct, err := GenerateNodeStruct(node)
	if err != nil {
		return
	}
	fmt.Println(nodeStruct)

	err = RegisterNode(node)
	if err != nil {
		return
	}
}
