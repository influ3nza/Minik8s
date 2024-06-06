package dns

import (
	"fmt"
	"minik8s/pkg/dns/dns_op"
	"minik8s/pkg/etcd"
	"time"
)

func InitDnsService() *dns_op.DnsService {
	srv := &dns_op.DnsService{}

	client, err := etcd.CreateEtcdInstance([]string{"http://192.168.1.13:2379"}, 5*time.Second)
	if err != nil {
		fmt.Println(err.Error())
	}
	srv.EtcdClient = client
	return srv
}
