package config

import (
	"time"
)

const (
	API_server_prefix string = "http://127.0.0.1:50000"

	API_get_nodes string = "/nodes"
	API_get_node  string = "/nodes/:name"
	API_add_node  string = "/nodes/add"

	API_update_pod string = "/pods/update"
	API_add_pod    string = "/pods/add"

	API_add_service string = "/services/add"
)

type ServerConfig struct {
	Port          int32
	TrustedProxy  []string
	EtcdEndpoints []string
	EtcdTimeout   time.Duration
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Port:          50000,
		TrustedProxy:  []string{"127.0.0.1"},
		EtcdEndpoints: []string{"localhost:2379"},
		EtcdTimeout:   5 * time.Second,
	}
}
