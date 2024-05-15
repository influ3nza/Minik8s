package pod_manager

import (
	"fmt"
	"minik8s/pkg/kubelet/util"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func GetPodIp(namespace string, containerName string) (string, error) {
	res, err := util.GetContainerInfo(namespace, ".NetworkSettings.IPAddress", containerName)
	if err != nil {
		fmt.Println("Failed At GetPodIp line 11 ", err.Error())
		return "", err
	}

	res = strings.TrimSuffix(res, "\n")
	return res, nil
}

func GetPodPid(namespace string, containerName string) (int, error) {
	res, err := util.GetContainerInfo(namespace, ".State.Pid", containerName)
	if err != nil {
		fmt.Println("Failed At GetPodPid line 21", err.Error())
		return 0, err
	}
	if strings.HasSuffix(res, "\n") {
		res = strings.TrimRight(res, "\n")
	}

	pid, err_ := strconv.Atoi(res)
	if err_ != nil {
		fmt.Println("Failed At GetPodPid line 27 ", err_.Error())
		return 0, err_
	}

	return pid, nil
}

func GetPodNetConfFile(namespace string, container string) ([]string, error) {
	dirPath := filepath.Join("./", namespace, container)
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return []string{}, err
	}
	_, err = util.CpContainer(namespace, container, "/etc/resolv.conf", fmt.Sprintf("%s/reslov.conf", dirPath), true)
	if err != nil {
		fmt.Println("GetPodNetConfFile Failed At line 20", err.Error())
		return []string{}, err
	}
	_, err = util.CpContainer(namespace, container, "/etc/hosts", fmt.Sprintf("%s/hosts", dirPath), true)
	if err != nil {
		fmt.Println("GetPodNetConfFile Failed At line 26 ", err.Error())
		return []string{}, err
	}
	return []string{fmt.Sprintf("%s/reslov.conf", dirPath), fmt.Sprintf("%s/hosts", dirPath)}, nil

}

func GenPodNetConfFile(namespace string, container string, pauseId string) error {
	dirPath := filepath.Join("./", namespace, pauseId)
	_, err := util.CpContainer(namespace, container, "/etc/resolv.conf", fmt.Sprintf("%s/reslov.conf", dirPath), false)
	if err != nil {
		fmt.Println("GenPodNetConfFile Failed At line 36 ", err.Error())
		return err
	}
	_, err = util.CpContainer(namespace, container, "/etc/hosts", fmt.Sprintf("%s/hosts", dirPath), false)
	if err != nil {
		fmt.Println("GenPodNetConfFile Failed At line 41 ", err.Error())
		return err
	}
	return nil
}

func RmLocalFile(files []string) error {
	for _, file := range files {
		err := os.Remove(file)
		if err != nil {
			fmt.Println("RmLocal File At line 32 ", err.Error())
		}
	}
	return nil
}
