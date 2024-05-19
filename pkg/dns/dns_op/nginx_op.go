package dns_op

import (
	"fmt"
	"os"
	"text/template"
)

const nginxConfigFile = "/mydata/nginx/nginx.conf"

var DNSRules AllDNSes

func RewriteNginx() error {
	file, err := os.OpenFile(nginxConfigFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("open Nginx File Failed, %s", err.Error())
	}
	writeString, err := file.WriteString(`worker_processes  5;  ## Default: 1
	error_log  ./error.log debug;
	pid        ./nginx.pid;
	worker_rlimit_nofile 8192;
	
	events {
	  worker_connections  4096;  ## Default: 1024
	}`)
	if err != nil {
		return fmt.Errorf("write Nginx Config Failed, %s", err.Error())
	} else {
		fmt.Println(writeString)
	}

	tmpl := template.Must(template.ParseFiles("/GJX/minik8s/pkg/dns/dns_op/nginx.tmpl"))
	res := tmpl.Execute(file, DNSRules)
	if res != nil {
		return fmt.Errorf("write To Nginx Config File Failed, %s", res.Error())
	}
	return nil
}
