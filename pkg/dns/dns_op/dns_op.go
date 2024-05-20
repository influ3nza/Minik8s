package dns_op

import (
	"encoding/json"
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/etcd"
	"os/exec"
	"strconv"
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
	etcdStore := ParseDns(dns.Host)
	value := map[string]string{
		"host": "192.168.1.13",
	}
	str, _ := json.Marshal(value)
	err := d.EtcdClient.Put(etcdStore, str)
	if err != nil {
		return fmt.Errorf("add Dns To Etcd Failed, %s", err.Error())
	}
	server := Server{
		Port:        "80",
		Domain:      dns.Host,
		ProxyPasses: []ProxyPass{},
	}
	for _, path := range dns.Paths {
		proxyPass := ProxyPass{
			Path: path.SubPath,
			IP:   path.ServiceIp,
			Port: strconv.Itoa(int(path.Port)),
		}
		server.ProxyPasses = append(server.ProxyPasses, proxyPass)
	}
	DNSRules.Servers = append(DNSRules.Servers, server)
	err = RewriteNginx()
	if err != nil {
		return fmt.Errorf("rewrite Nginx Failed In Add Dns, %s", err.Error())
	}

	err = RestartNginx()
	if err != nil {
		return fmt.Errorf("restart Nginx Failed In Add Dns, %s", err.Error())
	}
	return nil
}

func (d *DnsService) DeleteDns(dns *api_obj.Dns) error {
	key := ParseDns(dns.Host)
	err := d.EtcdClient.Del(key)
	if err != nil {
		fmt.Println("Delete Dns Failed In Del Etcd, ", err.Error())
	}
	for idx, rule := range DNSRules.Servers {
		if dns.Host == rule.Domain {
			DNSRules.Servers = append(DNSRules.Servers[:idx], DNSRules.Servers[idx+1:]...)
		}
	}
	err = RewriteNginx()
	if err != nil {
		return fmt.Errorf("reweite Nginx Failed In Del Dns, %s", err.Error())
	}

	err = RestartNginx()
	if err != nil {
		return fmt.Errorf("restart Nginx Failed In Del Dns, %s", err.Error())
	}
	return nil
}

func ParseDns(domain string) string {

	fullPath := reverseDomain(domain)
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
