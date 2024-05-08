package kube_proxy

import (
	"fmt"
	"github.com/vishvananda/netlink/nl"
	"golang.org/x/sys/unix"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/message"
	"net"
	"os/exec"
	"strconv"

	"github.com/moby/ipvs"
)

type ProxyManager struct {
	Consumer    *message.MsgConsumer
	IpvsHandler *ipvs.Handle
	Services    map[string]*MainService
}

func InitManager() *ProxyManager {
	consumer, err := message.NewConsumer("proxy", "proxy")
	if err != nil {
		fmt.Println("Error At Proxy line 25 ", err.Error())
		return nil
	}

	handler, err := ipvs.New("")
	if err != nil {
		fmt.Println("Error At Proxy line 31 ", err.Error())
		return nil
	}

	services := map[string]*MainService{}

	manager := &ProxyManager{
		Consumer:    consumer,
		IpvsHandler: handler,
		Services:    services,
	}
	return manager
}

func (m *ProxyManager) CreateService(srv *api_obj.Service) error {
	if srv.MetaData.UUID == "" {
		return fmt.Errorf("Error UUID is NULL")
	}
	mainService := &MainService{
		Srv: map[string]*Service{},
	}
	for _, miniSrv := range srv.Spec.Ports {
		label := fmt.Sprintf("%s:%s", srv.Spec.ClusterIP, miniSrv.Port)
		srcPort, _ := strconv.Atoi(miniSrv.Port)
		ipvsSrv := &ipvs.Service{
			Address:       net.ParseIP(srv.Spec.ClusterIP),
			Protocol:      unix.IPPROTO_TCP,
			Port:          uint16(srcPort),
			SchedName:     ipvs.RoundRobin,
			Netmask:       0xFFFFFFFF,
			AddressFamily: nl.FAMILY_V4,
		}

		err := m.IpvsHandler.NewService(ipvsSrv)
		if err != nil {
			fmt.Println("Create Service Failed At line 71 ", err.Error())
			return err
		}
		args := []string{"-t", "nat", "-A", "POSTROUTING", "-m", "ipvs", "--vaddr", srv.Spec.ClusterIP, "--vport", miniSrv.Port, "-j", "MASQUERADE"}
		_, err = exec.Command("iptables", args...).CombinedOutput()
		if err != nil {
			fmt.Println("Failed Add iptables At line 79 ", err.Error())
		}

		mySrv := &Service{
			Service:   ipvsSrv,
			EndPoints: map[string]*ipvs.Destination{},
		}
		mainService.Srv[label] = mySrv
	}

	m.Services[srv.MetaData.UUID] = mainService
	args := []string{"addr", "add", srv.Spec.ClusterIP + "/24", "dev", "flannel.1"}
	_, err := exec.Command("ip", args...).CombinedOutput()
	if err != nil {
		fmt.Println("Error Bind Net At Line 88 ", err.Error())
	}

	return err
}

func (m *ProxyManager) DelService(uuid string, ip string) error {
	mainSrc := m.Services[uuid]
	if mainSrc == nil {
		return fmt.Errorf("Error No Such Service")
	}

	var e error
	for _, srv := range mainSrc.Srv {
		if realIp := srv.Service.Address.String(); realIp != ip {
			return fmt.Errorf("real srv ip %s is not equal to ip %s", realIp, ip)
		}

		if err := m.delService(srv); err != nil {
			e = err
			fmt.Println("Del Srv Failed ", err.Error())
		}
		args := []string{"-t", "nat", "-D", "POSTROUTING", "-m", "ipvs", "--vaddr", ip, "--vport", strconv.Itoa(int(srv.Service.Port)), "-j", "MASQUERADE"}
		_, err := exec.Command("iptables", args...).CombinedOutput()
		if err != nil {
			fmt.Println("Failed Add iptables At line 107 ", err.Error())
		}
	}

	args := []string{"addr", "del", ip + "/24", "dev", "flannel.1"}
	_, err := exec.Command("ip", args...).CombinedOutput()
	if err != nil {
		fmt.Println("Error Bind Net At Line 109 ", err.Error())
	}

	delete(m.Services, uuid)
	return e
}

func (m *ProxyManager) delService(srv *Service) error {
	for _, dst := range srv.EndPoints {
		err := m.IpvsHandler.DelDestination(srv.Service, dst)
		if err != nil {
			fmt.Println("Err occurred At DelService Line 102 ", err.Error())
		}
	}

	err := m.IpvsHandler.DelService(srv.Service)
	if err != nil {
		fmt.Println("Err occurred At DelService line 108 ", err.Error())
		return err
	}

	return nil
}

//func (m *ProxyManager) AddEndPoint(ep *api_obj.Endpoint) error {
//
//}
//
//func (m *ProxyManager) DelEndPoint(ep *api_obj.Endpoint) error {
//
//}
