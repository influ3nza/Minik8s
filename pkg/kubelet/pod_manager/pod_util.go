package pod_manager

import (
	"fmt"
	"minik8s/pkg/kubelet/util"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// GetPodIp 获取pod的ip
/*
 * 参数
 *  namespace: string 命名空间
 *  containerName: string 容器名称
 *
 * 返回
 *  string: ip地址
 *  error: 错误信息
 */
func GetPodIp(namespace string, containerName string) (string, error) {
	res, err := util.GetContainerInfo(namespace, ".NetworkSettings.IPAddress", containerName)
	if err != nil {
		fmt.Println("Failed At GetPodIp line 11 ", err.Error())
		return "", err
	}

	res = strings.TrimSuffix(res, "\n")
	return res, nil
}

// GetPodPid 获取pod的pid
/*
 * 参数
 *  namespace: string 命名空间
 *  containerName: string 容器名称
 *
 * 返回
 *  int: pid
 *  error: 错误信息
 */
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

// GetPodNetConfFile 获取pod的网络配置文件
/*
 * 参数
 *  namespace: string 命名空间
 *  container: string 容器名称
 *
 * 返回
 *  []string: 文件路径
 *  error: 错误信息
 */
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

	content := "nameserver 192.168.1.13\n"
	reslovPath := fmt.Sprintf("%s/reslov.conf", dirPath)
	data, err := os.ReadFile(reslovPath)
	if err != nil {
		fmt.Println("Write in reslov file dns server failed at line 59 ", err.Error())
		return []string{}, err
	}

	data = append([]byte(content), data...)

	err = os.WriteFile(reslovPath, data, os.ModePerm)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return []string{}, err
	}

	_, err = util.CpContainer(namespace, container, "/etc/hosts", fmt.Sprintf("%s/hosts", dirPath), true)
	if err != nil {
		fmt.Println("GetPodNetConfFile Failed At line 26 ", err.Error())
		return []string{}, err
	}
	return []string{fmt.Sprintf("%s/reslov.conf", dirPath), fmt.Sprintf("%s/hosts", dirPath)}, nil

}

// GenPodNetConfFile 生成pod的网络配置文件
/*
 * 参数
 *  namespace: string 命名空间
 *  container: string 容器名称
 *  pauseId: string pause容器id
 *
 * 返回
 *  error: 错误信息
 */
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

// RmLocalFile 删除本地文件
/*
 * 参数
 *  files: []string 文件路径
 *
 * 返回
 *  error: 错误信息
 */
func RmLocalFile(files []string) error {
	for _, file := range files {
		err := os.Remove(file)
		if err != nil {
			fmt.Println("RmLocal File At line 32 ", err.Error())
		}
	}
	return nil
}
