package main

import (
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/dns/dns_op"
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
	err := dns_op.RewriteNginx(dns1)
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
	// fmt.Println(string(jsonData))
	err = dns_op.RewriteNginx(dns2)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(dns_op.DNSRules)
}
