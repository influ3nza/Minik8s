package function

import (
	"encoding/json"
	"fmt"
	"minik8s/pkg/etcd"
	"minik8s/pkg/kubelet/pod_manager"
	"os/exec"
	"time"
)

func StartRegistry() (string, error) {
	etcdClient, err := etcd.CreateEtcdInstance([]string{"192.168.1.13:2379"}, 5*time.Second)
	if err != nil {
		return "", fmt.Errorf("start Registry Failed %s", err.Error())
	}
	cmd := []string{"run", "-d", "--net", "flannel", "-v", "/mydata/docker/registry/:/var/lib/registry", "--name", "my-registry", "registry:latest"}
	_, err = exec.Command("nerdctl", cmd...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("start Registry Failed %s", err.Error())
	}
	ip, err := pod_manager.GetPodIp("default", "my-registry")
	key := "/savedns/io/my-registry"
	value := map[string]string{
		"host": ip,
	}
	str, _ := json.Marshal(value)
	err = etcdClient.Put(key, str)
	if err != nil {
		return "", fmt.Errorf("start Registry Failed %s", err)
	}

	return ip, nil
}

func CreateImage() error {
	return nil
}
