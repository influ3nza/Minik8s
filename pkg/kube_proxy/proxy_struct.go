package kube_proxy

import "github.com/moby/ipvs"

type Service struct {
	Service   *ipvs.Service
	EndPoints map[string]*ipvs.Destination
	NodePort  int32
}

type MainService struct {
	Srv map[string]*Service
}
