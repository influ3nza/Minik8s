package function

import (
	"encoding/json"
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/etcd"
	"minik8s/pkg/kubelet/pod_manager"
	"os/exec"
	"strings"
	"time"
)

var serverDns string = "my-register.io"

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

func CreateImage(function *api_obj.Function) error {
	path := "/mydata/" + function.Metadata.UUID + "/"
	cmd := exec.Command("nerdctl", "build", "-t", function.Metadata.Name, path)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("create Image Failed %s", err.Error())
	}
	cmd = exec.Command("nerdctl", "tag", function.Metadata.Name, serverDns+":5000/"+function.Metadata.Name+":latest")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("tag Image Failed %s", err.Error())
	}

	imgName := serverDns + ":5000/" + function.Metadata.Name + ":latest"
	cmd = exec.Command("nerdctl", "push", "--insecure-registry", imgName)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("push Image Failed %s", err.Error())
	}

	return nil
}

func FindImage(name string) bool {
	cmd := exec.Command("nerdctl", "images", name)
	opt, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("find Local Image Failed %s", err.Error())
		return false
	}
	result := strings.TrimSpace(string(opt))
	fmt.Printf("find Image Res is %s", result)
	if strings.Contains(result, name) {
		return true
	} else {
		return false
	}
}

func DeleteImage(name string) error {
	imgName := serverDns + ":5000/" + name + ":latest"
	if ok := FindImage(imgName); ok {
		cmd := exec.Command("nerdctl", "rmi", imgName)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("delete Image Failed %s", err.Error())
		}
	}

	return nil
}
