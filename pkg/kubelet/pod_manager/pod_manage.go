package pod_manager

import (
	"fmt"
	"github.com/containerd/containerd"
	"minik8s/pkg/api_obj"
)

// todo 需要Master上有DNS服务器
func AddPod(pod *api_obj.Pod) error {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		fmt.Println("Create Client Failed At")
		return err
	}

}
