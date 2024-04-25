package util

import (
	"fmt"
	"os/exec"
)

var path, _ = exec.LookPath("nerdctl")

const FirstSandbox = "registry.aliyuncs.com/google_containers/pause:latest"

func Exec(namespace string, args ...string) (string, error) {
	str := []string{"-n", namespace}
	str = append(str, args...)
	res, err := exec.Command(path, str...).CombinedOutput()
	// fmt.Println(string(res), err)
	return string(res), err
}

func PrintCmd(namespace string, args ...string) string {
	str := []string{"-n", namespace}
	str = append(str, args...)
	fmt.Println(str)
	return str[0]
}

func Mkdir(path string) (string, error) {
	res, err := exec.Command("mkdir", "-p", path).CombinedOutput()
	fmt.Println("At Mkdir line 22", string(res), err)
	return string(res), err
}

func CpContainer(namespace string, containerId string, containerPath string, hostPath string, inOrOut bool) (string, error) {
	if inOrOut == true {
		cmd := []string{"cp", containerId + ":" + containerPath, hostPath}
		res, err := Exec(namespace, cmd...)
		PrintCmd(namespace, cmd...)
		if err != nil {
			fmt.Println("Err At GetFileFromContainer 39 ", err.Error())
			return "", err
		}
		return res, nil
	} else {
		cmd := []string{"cp", hostPath, containerId + ":" + containerPath}
		res, err := Exec(namespace, cmd...)
		PrintCmd(namespace, cmd...)
		if err != nil {
			fmt.Println("Err At GetFileFromContainer 48 ", err.Error())
			return "", err
		}
		return res, nil
	}
}

func RunContainer(namespace string, name string) (string, error) {
	cmd := []string{"run", "-d", "--name", name, "--net", "flannel", "--label", "podStart=pause4pod", FirstSandbox}
	PrintCmd(namespace, cmd...)
	// fmt.Println("RunContainer")
	res, err := Exec(namespace, cmd...)
	if err != nil {
		return "", err
	}
	return res, nil
	//res := PrintCmd(namespace, cmd...)
	//return res, nil
}

func GetContainerInfo(namespace string, info string, containerId string) (string, error) {
	cmd := []string{"inspect", "-f", fmt.Sprintf("{{%s}}", info), containerId}
	res, err := Exec(namespace, cmd...)
	if err != nil {
		return "", err
	}
	return res, nil
}

func RmForce(namespace string, name string) (string, error) {
	cmd := []string{"rm -f", name}
	PrintCmd(namespace, cmd...)
	res, err := Exec(namespace, cmd...)
	if err != nil {
		return "", err
	}
	return res, nil
}

func StopContainer(namespace string, name string) (string, error) {
	cmd := []string{"stop", name}
	PrintCmd(namespace, cmd...)
	res, err := Exec(namespace, cmd...)
	if err != nil {
		return "", err
	}
	return res, nil
}

func RemoveContainer(namespace string, name string) (string, error) {
	cmd := []string{"rm", name}
	res, err := Exec(namespace, cmd...)
	if err != nil {
		return "", err
	}
	return res, nil
}
