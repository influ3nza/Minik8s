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
			ClusterIP: "10.20.0.1",
			Ports: []api_obj.ServicePort{
				{
					Name:       "testname",
					Protocol:   "tcp",
					Port:       "80",
					TargetPort: "80",
				}, {
					Name:       "testname",
					Protocol:   "tcp",
					Port:       "890",
					TargetPort: "890",
				},
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
	err_ := manager.DelService(serv.MetaData.UUID, serv.Spec.ClusterIP)
	if err_ != nil {
		fmt.Println("delete failed at main line 73, ", err_.Error())
		return
	}
	return
}
