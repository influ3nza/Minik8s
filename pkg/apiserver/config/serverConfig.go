package config

import (
	"time"
)

type ServerConfig struct {
	Port          int32
	EtcdEndpoints []string
	EtcdTimeout   time.Duration
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Port:          8080,
		EtcdEndpoints: []string{"localhost:2379"},
		EtcdTimeout:   5 * time.Second,
	}
}
