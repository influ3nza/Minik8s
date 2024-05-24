package monitor

import (
	"encoding/json"
	"fmt"
	"minik8s/pkg/api_obj"
	"net/http"
	"strconv"
	"strings"
)

func GenerateNodeStruct(node *api_obj.Node) (*ConsulConfig, error) {
	hostname := node.NodeMetadata.Name

	Ip := node.NodeStatus.Addresses.InternalIp

	config := &ConsulConfig{
		Id:      "node-exporter-" + hostname,
		Name:    "node-exporter-" + Ip,
		Tags:    []string{"node"},
		Address: Ip,
		Port:    9100,
		Meta: map[string]string{
			"app":  "node",
			"team": "minik8s fourth group",
		},
		Checks: []Check{
			{
				Http:     "http://" + Ip + "9100/metrics",
				Interval: "15s",
			},
		},
	}

	return config, nil
}

func RegisterNode(node *api_obj.Node) error {
	cfg, err := GenerateNodeStruct(node)
	if err != nil {
		fmt.Println("Err Register Node ", err.Error())
		return err
	}
	targetUrl := "http://192.168.1.13:8500/v1/agent/service/register"
	jsonCfg, err := json.Marshal(*cfg)
	payLoad := strings.NewReader(string(jsonCfg))
	req, err := http.NewRequest(http.MethodPut, targetUrl, payLoad)
	if err != nil {
		fmt.Println("Create PUT RegisterNode Failed, ", err.Error())
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	_, err = http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println("Send PUT RegisterNode Failed, ", err.Error())
		return err
	}

	return nil
}

func UnRegisterNode(hostname string) error {
	key := "node-exporter-" + hostname
	targetUrl := "http://192.168.1.13:8500/v1/agent/service/deregister" + key
	payLoad := strings.NewReader("")
	req, err := http.NewRequest(http.MethodPut, targetUrl, payLoad)
	if err != nil {
		fmt.Println("Create PUT UnregisterNode Failed, ", err.Error())
		return err
	}

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Send PUT UnregisterNode Failed, ", err.Error())
		return err
	}

	return nil
}

func UnRegisterPod(namespace string, name string) error {
	key := namespace + "-" + name
	targetUrl := "http://192.168.1.13:8500/v1/agent/service/deregister/" + key
	payLoad := strings.NewReader("")
	req, err := http.NewRequest(http.MethodPut, targetUrl, payLoad)
	if err != nil {
		fmt.Println("Create PUT UnregisterPod Failed, ", err.Error())
		return err
	}

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Send PUT UnregisterNode Failed, ", err.Error())
		return err
	}

	return nil
}

func RegisterPod(pod *api_obj.Pod) error {
	serverIp := pod.PodStatus.PodIP
	port, err := strconv.Atoi(pod.MetaData.Labels["metricsPort"])

	if err != nil {
		return fmt.Errorf("atoi Failed, %s", err.Error())
	}

	config := &ConsulConfig{
		Id:      pod.MetaData.NameSpace + "-" + pod.MetaData.Name,
		Name:    pod.MetaData.NameSpace + "-" + pod.MetaData.Name,
		Tags:    []string{"pod"},
		Address: serverIp,
		Port:    int32(port),
		Meta: map[string]string{
			"app": "pod",
		},
		Checks: []Check{
			{
				Http:     "http://" + serverIp + ":" + fmt.Sprintf("%d", port) + "/metrics",
				Interval: "15s",
			},
		},
	}

	targetUrl := "https://192.168.1.13:8500/v1/agent/service/register"
	jsonCfg, err := json.Marshal(*config)
	payLoad := strings.NewReader(string(jsonCfg))
	req, err := http.NewRequest(http.MethodPut, targetUrl, payLoad)
	if err != nil {
		fmt.Println("Create PUT RegisterPod Failed, ", err.Error())
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	_, err = http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println("Send PUT RegisterPod Failed, ", err.Error())
		return err
	}

	return nil
}
