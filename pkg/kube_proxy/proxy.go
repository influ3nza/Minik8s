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
		return fmt.Errorf("error UUID is NULL")
	}
	mainService := &MainService{
		Srv: map[string]*Service{},
	}
	for idx, miniSrv := range srv.Spec.Ports {
		label := fmt.Sprintf("%s:%d", srv.Spec.ClusterIP, miniSrv.Port)
		ipvsSrv := &ipvs.Service{
			Address:       net.ParseIP(srv.Spec.ClusterIP),
			Protocol:      unix.IPPROTO_TCP,
			Port:          uint16(miniSrv.Port),
			SchedName:     ipvs.RoundRobin,
			Netmask:       0xFFFFFFFF,
			AddressFamily: nl.FAMILY_V4,
		}

		err := m.IpvsHandler.NewService(ipvsSrv)
		if err != nil {
			fmt.Println("Create Service Failed At line 66 ", err.Error())
			return err
		}
		args := []string{"-t", "nat", "-A", "POSTROUTING", "-m", "ipvs", "--vaddr", srv.Spec.ClusterIP, "--vport", strconv.Itoa(int(miniSrv.Port)), "-j", "MASQUERADE"}
		_, err = exec.Command("iptables", args...).CombinedOutput()
		if err != nil {
			fmt.Println("Failed Add iptables At line 70 ", err.Error())
		}

		args = []string{"-t", "nat", "-I", "POSTROUTING", "-m", "ipvs", "--vaddr", srv.Spec.ClusterIP, "--vport", strconv.Itoa(int(miniSrv.Port)), "-j", "MASQUERADE"}
		_, err = exec.Command("iptables", args...).CombinedOutput()
		if err != nil {
			fmt.Println("Failed Add iptables At line 76 ", err.Error())
		}
		fmt.Println(srv.Spec.Type)
		switch srv.Spec.Type {
		case api_obj.NodePort:
			{
				ip, _ := GetLocalIP()
				var nodePort int32 = 0
				if miniSrv.NodePort != 0 {
					nodePort = miniSrv.NodePort
				} else {
					p, _ := GetFreePort()
					nodePort = int32(p)
				}
				fmt.Printf("Select NodePort is %d", nodePort)
				clusterLabel := fmt.Sprintf("%s:%d", srv.Spec.ClusterIP, miniSrv.Port)
				fmt.Println("Label is ", clusterLabel)
				cmd := []string{"-t", "nat", "-A", "PREROUTING", "-p", "tcp", "-d", ip, "--dport", fmt.Sprintf("%d", nodePort), "-j", "DNAT", "--to-destination", clusterLabel}
				_, err = exec.Command("iptables", cmd...).CombinedOutput()
				fmt.Println("iptables ", cmd)
				if err != nil {
					fmt.Println("Failed Add iptables NodePort At line 91 ", err.Error())
				}
				mySrv := &Service{
					Service:   ipvsSrv,
					EndPoints: map[string]*ipvs.Destination{},
					NodePort:  nodePort,
				}
				srv.Spec.Ports[idx].NodePort = nodePort
				mainService.Srv[label] = mySrv
			}
		case api_obj.ClusterIP:
			{
				mySrv := &Service{
					Service:   ipvsSrv,
					EndPoints: map[string]*ipvs.Destination{},
					NodePort:  0,
				}
				mainService.Srv[label] = mySrv
			}
		default:
			{
				mySrv := &Service{
					Service:   ipvsSrv,
					EndPoints: map[string]*ipvs.Destination{},
					NodePort:  0,
				}
				mainService.Srv[label] = mySrv
			}
		}
		fmt.Println("Here")
	}

	m.Services[srv.MetaData.UUID] = mainService
	args := []string{"addr", "add", srv.Spec.ClusterIP + "/24", "dev", "flannel.1"}
	_, err := exec.Command("ip", args...).CombinedOutput()
	if err != nil {
		fmt.Println("Error Bind Net At Line 88 ", err.Error())
	}

	return err
}

func (m *ProxyManager) DelService(srv *api_obj.Service) error {
	mainSrc := m.Services[srv.MetaData.UUID]
	if mainSrc == nil {
		return fmt.Errorf("error No Such Service")
	}

	var e error
	for _, miniSrv := range mainSrc.Srv {
		if realIp := miniSrv.Service.Address.String(); realIp != srv.Spec.ClusterIP {
			return fmt.Errorf("real srv ip %s is not equal to ip %s", realIp, srv.Spec.ClusterIP)
		}

		if err := m.delService(miniSrv); err != nil {
			e = err
			fmt.Println("Del Srv Failed ", err.Error())
		}
		args := []string{"-t", "nat", "-D", "POSTROUTING", "-m", "ipvs", "--vaddr", srv.Spec.ClusterIP, "--vport", fmt.Sprintf("%d", miniSrv.Service.Port), "-j", "MASQUERADE"}
		for i := 0; i < 2; i++ {
			_, err := exec.Command("iptables", args...).CombinedOutput()
			fmt.Println("iptables ", args)
			if err != nil {
				fmt.Println("Failed Add iptables At line 150 ", err.Error())
			}
		}

		switch srv.Spec.Type {
		case api_obj.NodePort:
			{
				ip, _ := GetLocalIP()
				clusterLabel := fmt.Sprintf("%s:%d", srv.Spec.ClusterIP, miniSrv.Service.Port)
				args = []string{"-t", "nat", "-D", "PREROUTING", "-p", "tcp", "-d", ip, "--dport", fmt.Sprintf("%d", miniSrv.NodePort), "-j", "DNAT", "--to-destination", clusterLabel}
				_, err := exec.Command("iptables", args...).CombinedOutput()
				if err != nil {
					fmt.Println("Failed Add iptables At line 161 ", err.Error())
				}
			}
		case api_obj.ClusterIP:
		default:
		}
	}

	args := []string{"addr", "del", srv.Spec.ClusterIP + "/24", "dev", "flannel.1"}
	_, err := exec.Command("ip", args...).CombinedOutput()
	if err != nil {
		fmt.Println("Error Bind Net At Line 109 ", err.Error())
	}

	delete(m.Services, srv.MetaData.UUID)
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

func (m *ProxyManager) AddEndPoint(ep *api_obj.Endpoint) error {
	mainSrv := m.Services[ep.SrvUUID]
	if mainSrv == nil {
		return fmt.Errorf("no Such Service UUID %s", ep.SrvUUID)
	}

	label := fmt.Sprintf("%s:%d", ep.SrvIP, ep.SrvPort)
	realSrv := mainSrv.Srv[label]
	if realSrv == nil {
		return fmt.Errorf("no Such Service UUID %s ip:port %s", ep.SrvUUID, label)
	}

	if ep.Weight == 0 {
		ep.Weight = 1
	}

	dst := &ipvs.Destination{
		Address:       net.ParseIP(ep.PodIP),
		Port:          uint16(ep.PodPort),
		Weight:        ep.Weight,
		AddressFamily: nl.FAMILY_V4,
	}
	dstLabel := fmt.Sprintf("%s:%d", ep.PodIP, ep.PodPort)
	err := m.IpvsHandler.NewDestination(realSrv.Service, dst)
	if err != nil {
		return fmt.Errorf("create EndPoint Failed: %s", err.Error())
	}
	realSrv.EndPoints[dstLabel] = dst
	return nil
}

func (m *ProxyManager) DelEndPoint(ep *api_obj.Endpoint) error {
	mainSrv := m.Services[ep.SrvUUID]
	if mainSrv == nil {
		return fmt.Errorf("no Such Service UUID %s", ep.SrvUUID)
	}

	label := fmt.Sprintf("%s:%d", ep.SrvIP, ep.SrvPort)
	realSrv := mainSrv.Srv[label]
	if realSrv == nil {
		return fmt.Errorf("no Such Service UUID %s ip:port %s", ep.SrvUUID, label)
	}

	dstLabel := fmt.Sprintf("%s:%d", ep.PodIP, ep.PodPort)
	dst := realSrv.EndPoints[dstLabel]
	if dst == nil {
		return fmt.Errorf("no Such EndPoint Srv is %s, Ep is %s", label, dstLabel)
	}

	err := m.IpvsHandler.DelDestination(realSrv.Service, dst)
	if err != nil {
		return fmt.Errorf("del EndPoint Failed, Err is %s", err.Error())
	}

	delete(realSrv.EndPoints, dstLabel)
	return nil
}
