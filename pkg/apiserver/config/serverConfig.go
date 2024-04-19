package config

import (
	"time"
)

type ServerConfig struct {
	Port          int32
	TrustedProxy  []string
	EtcdEndpoints []string
	EtcdTimeout   time.Duration
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Port:          8080,
		TrustedProxy:  []string{"127.0.0.1"},
		EtcdEndpoints: []string{"localhost:2379"},
		EtcdTimeout:   5 * time.Second,
	}
}
