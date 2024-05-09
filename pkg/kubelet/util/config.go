package util

const Cpu string = "2"
const Memory string = "8Gi"
const IpAddress string = "192.168.1.13"
const Port string = "20000"
const ApiServer string = "http://10.119.13.178:50000"

type KubeConfig struct {
	ApiServer string            `yaml:"ApiServer"`
	Ip        string            `yaml:"Ip"`
	Port      int32             `yaml:"Port"`
	TotalCpu  string            `yaml:"TotalCpu"`
	TotalMem  string            `yaml:"TotalMem"`
	Label     map[string]string `yaml:"Label"`
}
