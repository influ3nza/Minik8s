package dns

import (
	"fmt"
	"minik8s/pkg/api_obj"
	"os/exec"
	"strings"
)

func StartDns(configPath string) error {
	cmd := []string{"-conf", configPath}
	res, err := exec.Command("coredns", cmd...).CombinedOutput()
	if err != nil {
		fmt.Println("Start Dns Server Failed ", err.Error())
		return err
	}
	fmt.Println("Start Dns ", res)
	return nil
}

func AddDns(dns *api_obj.Dns) error {
	for
}

func DeleteDns(dns *api_obj.Dns) error {

}

func parseDns(host string, path string) string {
	if strings.HasSuffix(host, ".") {
		host = strings.TrimSuffix(host, ".")
	}

	fullPath := fmt.Sprintf("%s.%s", host, path)
	return fullPath
}
