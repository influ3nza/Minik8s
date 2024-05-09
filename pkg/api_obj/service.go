package api_obj

import (
	"minik8s/pkg/api_obj/obj_inner"
)

type ServiceExportType string

const (
	NodePort     ServiceExportType = "NodePort"
	LoadBalancer ServiceExportType = "LoadBalancer"
	ExternalName ServiceExportType = "ExternalName"
	ClusterIP    ServiceExportType = "ClusterIP"
)

const (
	SERVICE_PENDING string = "service_pending"
	SERVICE_CREATED string = "service_created"
)

type ServiceStatus struct {
	Condition string     `json:"condition" yaml:"condition"`
	Endpoints []Endpoint `json:"endpoints" yaml:"endpoints"`
}

type ServicePort struct {
	Name       string `json:"name" yaml:"name"`
	Protocol   string `json:"protocol" yaml:"protocol"`
	Port       string `json:"port" yaml:"port"`
	TargetPort string `json:"targetPort" yaml:"targetPort"`
}

type ServiceSpec struct {
	Ports     []ServicePort     `json:"ports" yaml:"ports"`
	Selector  map[string]string `json:"selector" yaml:"selector"`
	ClusterIP string            `json:"clusterIP" yaml:"clusterIP"`
	Type      ServiceExportType `json:"type" yaml:"type"`
}

type Service struct {
	ApiVersion string               `json:"apiVersion" yaml:"apiVersion"`
	Kind       string               `json:"kind" yaml:"kind"`
	MetaData   obj_inner.ObjectMeta `json:"metadata" yaml:"metadata"`
	Spec       ServiceSpec          `json:"spec" yaml:"spec"`
	Status     ServiceStatus        `json:"status" yaml:"status"`
}
