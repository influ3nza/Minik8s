package monitor

type ConsulConfig struct {
	Id      string
	Name    string
	Tags    []string
	Address string
	Port    int32
	Meta    map[string]string
	Checks  []Check
}

type Check struct {
	Http     string
	Interval string
}
