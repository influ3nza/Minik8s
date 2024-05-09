package config

import (
	"time"
)

const (
	API_server_prefix string = "http://127.0.0.1:50000"

	API_get_nodes       string = "/nodes/getAll"
	API_get_node_prefix string = "/nodes/"
	API_get_node        string = "/nodes/:namespace/:name"
	API_add_node        string = "/nodes/add"

	API_update_pod                   string = "/pods/update"
	API_add_pod                      string = "/pods/add"
	API_get_pods                     string = "/pods/getAll"
	API_get_pods_by_node_prefix      string = "/pods/getByNode"
	API_get_pods_by_node             string = "/pods/getByNode/:nodename"
	API_get_pods_by_namespace_prefix string = "/pods/getByNamespace"
	API_get_pods_by_namespace        string = "/pods/getByNamespace/:namespace"

	API_add_service  string = "/services/add"
	API_get_services string = "/services/getAll"

	API_add_endpoint string = "/endpoints/add"
	//所有endpoint的名字{srvname}-{podname}
	API_delete_endpoints_prefix string = "/endpoints/deleteBatch/"
	API_delete_endpoints        string = "/endpoints/deleteBatch/:namespace/:srvname"
	API_delete_endpoint_prefix  string = "/endpoints/delete/"
	API_delete_endpoint         string = "/endpoints/delete/:namespace/:name"
)

type ServerConfig struct {
	Port          int32
	TrustedProxy  []string
	EtcdEndpoints []string
	EtcdTimeout   time.Duration
	MaxNodeCount  int32
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Port:          50000,
		TrustedProxy:  []string{"127.0.0.1"},
		EtcdEndpoints: []string{"localhost:2379"},
		EtcdTimeout:   5 * time.Second,
		MaxNodeCount:  10,
	}
}
