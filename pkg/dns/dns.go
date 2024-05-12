package dns

import (
	"fmt"
	"minik8s/pkg/dns/dns_op"
	"minik8s/pkg/etcd"
	"time"
)

func InitDnsService(filePath string) *dns_op.DnsService {
	srv := &dns_op.DnsService{
		ConfigFile: filePath,
	}

	err := srv.StartDns(srv.ConfigFile)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	client, err := etcd.CreateEtcdInstance([]string{"http://192.168.1.13:2379"}, 5*time.Second)
	if err != nil {
		fmt.Println(err.Error())
	}
	srv.EtcdClient = client
	return nil
}
