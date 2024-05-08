package kube_proxy

import (
	"fmt"
	"github.com/vishvananda/netlink/nl"
	"golang.org/x/sys/unix"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/message"
	"net"
	"strconv"

	"github.com/moby/ipvs"
)

type ProxyManager struct {
	Consumer    *message.MsgConsumer
	IpvsHandler *ipvs.Handle
	Services    map[string]*MainService
}

type Manager interface {
	CreateService(srv *api_obj.Service)
	DeleteService(uuid string, ip string, port int)
	AddEndPoint(ep *api_obj.Endpoint)
	DelEndPoint(ip string, port int)
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

		mySrv := &Service{
			Service:   ipvsSrv,
			EndPoints: map[string]*ipvs.Destination{},
		}
		mainService.Srv[label] = mySrv
	}

	m.Services[srv.MetaData.UUID] = mainService
	return nil
}

func (m *ProxyManager) DelService(uuid string, ip string, port int) error {
	mainSrc := m.Services[uuid]
	if mainSrc == nil {
		return fmt.Errorf("Error No Such Service")
	}

	label := fmt.Sprintf("%s:%d", ip, port)
	service := mainSrc.Srv[label]
	if service == nil {
		return fmt.Errorf("Error No Such Service")
	}

	for _, dst := range service.EndPoints {
		err := m.IpvsHandler.DelDestination(service.Service, dst)
		if err != nil {
			fmt.Println("Err occurred At DelService Line 102 ", err.Error())
		}
	}

	err := m.IpvsHandler.DelService(service.Service)
	if err != nil {
		fmt.Println("Err occurred At DelService line 108 ", err.Error())
		return err
	}

	delete(mainSrc.Srv, label)
	return nil
}

func (m *ProxyManager) AddEndPoint(ep *api_obj.Endpoint) {

}
