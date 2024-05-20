package main

import (
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/dns/dns_op"
	"os/exec"
	"strconv"
)

func main() {
	domain := "dns.example.com"
	reversedDomain := dns_op.ParseDns(domain)
	fmt.Println("Reversed domain:", reversedDomain)

	dns1 := &api_obj.Dns{
		Kind:       "DNS",
		ApiVersion: "v1",
		MetaData:   obj_inner.ObjectMeta{Name: "dns-test1"},
		Host:       "node1.com",
		Paths: []api_obj.Path{
			{
				SubPath:     "path1",
				ServiceIp:   "127.1.1.10",
				ServiceName: "service1",
				Port:        8010,
			},
			{
				SubPath:     "path2",
				ServiceIp:   "127.1.1.11",
				ServiceName: "service2",
				Port:        8011,
			},
		},
	}

	server := dns_op.Server{
		Port:        "80",
		Domain:      dns1.Host,
		ProxyPasses: []dns_op.ProxyPass{},
	}
	for _, path := range dns1.Paths {
		proxyPass := dns_op.ProxyPass{
			Path: path.SubPath,
			IP:   path.ServiceIp,
			Port: strconv.Itoa(int(path.Port)),
		}
		server.ProxyPasses = append(server.ProxyPasses, proxyPass)
	}
	dns_op.DNSRules.Servers = append(dns_op.DNSRules.Servers, server)

	err := dns_op.RewriteNginx()
	if err != nil {
		fmt.Println(err.Error())
	}

	dns2 := &api_obj.Dns{
		Kind:       "DNS",
		ApiVersion: "v1",
		MetaData:   obj_inner.ObjectMeta{Name: "dns-test2"},
		Host:       "node2.com",
		Paths: []api_obj.Path{
			{
				SubPath:     "srv1",
				ServiceIp:   "127.1.1.12",
				ServiceName: "service3",
				Port:        31000,
			},
		},
	}

	server = dns_op.Server{
		Port:        "80",
		Domain:      dns2.Host,
		ProxyPasses: []dns_op.ProxyPass{},
	}
	for _, path := range dns2.Paths {
		proxyPass := dns_op.ProxyPass{
			Path: path.SubPath,
			IP:   path.ServiceIp,
			Port: strconv.Itoa(int(path.Port)),
		}
		server.ProxyPasses = append(server.ProxyPasses, proxyPass)
	}
	dns_op.DNSRules.Servers = append(dns_op.DNSRules.Servers, server)

	err = dns_op.RewriteNginx()
	if err != nil {
		fmt.Println(err.Error())
	}
	otp, err := exec.Command("cat", "/mydata/nginx/nginx.conf").CombinedOutput()
	if err != nil {
		fmt.Println("Cat File Error")
		return
	} else {
		fmt.Println(otp)
	}
	fmt.Printf("\n")
	dns_op.DNSRules.Servers = append(dns_op.DNSRules.Servers[:1], dns_op.DNSRules.Servers[1+1:]...)
	err = dns_op.RewriteNginx()
	otp, err = exec.Command("cat", "/mydata/nginx/nginx.conf").CombinedOutput()
	if err != nil {
		fmt.Println("Cat File Error")
		return
	} else {
		fmt.Println(otp)
	}
	fmt.Println(dns_op.DNSRules)

	err = dns_op.RestartNginx()
	if err != nil {
		fmt.Println("Error While Restart Nginx ", err.Error())
		return
	}
}
