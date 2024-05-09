package main

import (
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/kube_proxy"
	"os/exec"
)

func InitEnv() error {
	_, err := exec.Command("modprobe", "ip_vs").CombinedOutput()
	if err != nil {
		fmt.Println("Init Env Error ", err.Error())
		return err
	}

	_, err = exec.Command("sysctl", "net.ipv4.vs.conntrack=1").CombinedOutput()
	if err != nil {
		fmt.Println("sys conntrack Error ", err.Error())
		return err
	}
	return nil
}

func main() {
	if err := InitEnv(); err != nil {
		return
	}

	serv := &api_obj.Service{
		ApiVersion: "v1",
		Kind:       "service",
		MetaData: obj_inner.ObjectMeta{
			Name:        "test",
			NameSpace:   "test",
			Labels:      map[string]string{},
			Annotations: map[string]string{},
			UUID:        "TESTUUID",
		},
		Spec: api_obj.ServiceSpec{
			Type:      api_obj.NodePort,
			Selector:  map[string]string{},
			ClusterIP: "172.20.0.1",
			Ports: []api_obj.ServicePort{
				{
					Name:       "testname",
					Protocol:   "tcp",
					Port:       "7840",
					TargetPort: "7840",
					NodePort:   30000,
				},
				//}, {
				//	Name:       "testname",
				//	Protocol:   "tcp",
				//	Port:       "890",
				//	TargetPort: "890",
				//	NodePort:   0,
				//},
			},
		},
		Status: api_obj.ServiceStatus{},
	}

	manager := kube_proxy.InitManager()
	if manager == nil {
		fmt.Println("Manager is Nil")
		return
	}
	err := manager.CreateService(serv)
	if err != nil {
		fmt.Println("Create Service Failed <.>")
	}

	res, _ := manager.IpvsHandler.GetServices()
	for _, r := range res {
		fmt.Println("Ip: ", r.Address, " Port: ", r.Port)
	}

	//out, err := exec.Command("ipvsadm", "-Ln").CombinedOutput()
	//fmt.Printf("%s\n", string(out))
	//
	//out, err = exec.Command("iptables-save").CombinedOutput()
	//fmt.Printf("%s\n", string(out))
	//
	//out, err = exec.Command("ip", "addr").CombinedOutput()
	//fmt.Printf("%s\n", string(out))

	ep := &api_obj.Endpoint{
		ApiVersion: "v1",
		Kind:       "ep",
		MetaData: obj_inner.ObjectMeta{
			Name:        "testep",
			NameSpace:   "test",
			Labels:      nil,
			Annotations: nil,
			UUID:        "",
		},
		SrvUUID: "TESTUUID",
		SrvIP:   "172.20.0.1",
		SrvPort: 7840,
		PodUUID: "UUID",
		PodIP:   "10.2.3.3",
		PodPort: 80,
		Weight:  2,
	}
	err = manager.AddEndPoint(ep)
	out, err := exec.Command("ipvsadm", "-Ln").CombinedOutput()
	fmt.Printf("%s\n", string(out))

	out, err = exec.Command("curl", "172.20.0.1:7840").CombinedOutput()
	fmt.Printf("%s\n", string(out))

	err = manager.DelEndPoint(ep)
	if err != nil {
		fmt.Println(err.Error())
	}
	out, err = exec.Command("ipvsadm", "-Ln").CombinedOutput()
	fmt.Printf("%s\n", string(out))

	//time.Sleep(60 * time.Second)
	err_ := manager.DelService(serv)
	if err_ != nil {
		fmt.Println("delete failed at main line 73, ", err_.Error())
		return
	}
	return
}
