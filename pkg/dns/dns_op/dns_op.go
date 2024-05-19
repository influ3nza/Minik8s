package dns_op

import (
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/etcd"
	"os/exec"
	"strings"
)

type DnsService struct {
	ConfigFile string
	EtcdClient *etcd.EtcdWrap
}

func (d *DnsService) StartDns(configPath string) error {
	cmd := []string{"-conf", configPath}
	res, err := exec.Command("coredns", cmd...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("start Dns Server Failed %s", err.Error())
	}
	fmt.Println("Start Dns ", res)
	return nil
}

func (d *DnsService) AddDns(dns *api_obj.Dns) error {
	for _, p := range dns.Paths {
		if p.EndPoint == nil {
			return fmt.Errorf("cannot generate dns, endpoint is null")
		}
		key := ParseDns(dns.Host, p.SubPath)
		value := p.EndPoint.SrvIP
		err := d.EtcdClient.Put(key, []byte(value))
		if err != nil {
			return fmt.Errorf("cannot generate dns, put etcd failed")
		}
	}
	return nil
}

func (d *DnsService) DeleteDns(dns *api_obj.Dns) error {
	for _, p := range dns.Paths {
		key := ParseDns(dns.Host, p.SubPath)
		err := d.EtcdClient.Del(key)
		if err != nil {
			return fmt.Errorf("cannot del dns, %s", err.Error())
		}
	}
	return nil
}

func ParseDns(host string, path string) string {
	if strings.HasSuffix(host, ".") {
		host = strings.TrimSuffix(host, ".")
	}

	fullPath := fmt.Sprintf("%s.%s", host, path)
	fullPath = reverseDomain(fullPath)
	fullPath = strings.Replace(fullPath, ".", "/", -1)
	fullPath = fmt.Sprintf("/savedns/%s", fullPath)
	fmt.Println(fullPath)
	return fullPath
}

func reverseDomain(domain string) string {
	parts := strings.Split(domain, ".")
	reversedParts := []string{}
	fmt.Println(parts, len(parts))
	// 反转域名的各个部分
	for i := len(parts) - 1; i >= 0; i-- {
		reversedParts = append(reversedParts, parts[i])
	}
	fmt.Println(reversedParts, len(reversedParts))
	// 组合反转后的域名
	reversedDomain := strings.Join(reversedParts, ".")

	return reversedDomain
}
