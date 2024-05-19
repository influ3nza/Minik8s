package main

import (
	"fmt"
	"minik8s/pkg/dns/dns_op"
	"strings"
)

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

func main() {
	domain := "dns.example.com"
	reversedDomain := dns_op.ParseDns("www", domain)
	fmt.Println("Reversed domain:", reversedDomain)
}
