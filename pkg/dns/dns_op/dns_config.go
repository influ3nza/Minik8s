package dns_op

type ProxyPass struct {
	Path string
	IP   string
	Port string
}

type Server struct {
	Port        string
	ServerName  string
	ProxyPasses []ProxyPass
}

type AllDNSes struct {
	Servers []Server
}
