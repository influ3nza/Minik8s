package monitor

import (
	"encoding/json"
	"fmt"
	"minik8s/pkg/api_obj"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func GetLocalIP() (ipv4 string, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, addr := range addrs {
		ipNet, isIpNet := addr.(*net.IPNet)
		if isIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String()
				return
			}
		}
	}
	return
}

func GenerateNodeStruct() (*ConsulConfig, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("get Hostname Failed, %s", err.Error())
	}

	localIp, err := GetLocalIP()
	if err != nil {
		return nil, fmt.Errorf("get Local Ip Failed, %s", err.Error())
	}

	config := &ConsulConfig{
		Id:      "node-exporter-" + hostname,
		Name:    "node-exporter-" + localIp,
		Tags:    []string{"node"},
		Address: localIp,
		Port:    9100,
		Meta: map[string]string{
			"app":  "node",
			"team": "minik8s fourth group",
		},
		Checks: []Check{
			{
				Http:     "http://" + localIp + "9100/metrics",
				Interval: "15s",
			},
		},
	}

	return config, nil
}

func RegisterNode() error {
	cfg, err := GenerateNodeStruct()
	if err != nil {
		fmt.Println("Err Register Node ", err.Error())
		return err
	}
	targetUrl := "https://192.168.1.13:8500/v1/agent/service/register"
	jsonCfg, err := json.Marshal(*cfg)
	payLoad := strings.NewReader(string(jsonCfg))
	req, err := http.NewRequest(http.MethodPut, targetUrl, payLoad)
	if err != nil {
		fmt.Println("Create PUT Failed, ", err.Error())
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	_, err = http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println("Send PUT Failed, ", err.Error())
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
				Http:     "http://" + serverIp + ":" + strconv.Itoa(port) + "/metrics",
				Interval: "15s",
			},
		},
	}

	targetUrl := "https://192.168.1.13:8500/v1/agent/service/register"
	jsonCfg, err := json.Marshal(*config)
	payLoad := strings.NewReader(string(jsonCfg))
	req, err := http.NewRequest(http.MethodPut, targetUrl, payLoad)
	if err != nil {
		fmt.Println("Create PUT Failed, ", err.Error())
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	_, err = http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println("Send PUT Failed, ", err.Error())
		return err
	}

	return nil
}
