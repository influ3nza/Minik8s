package monitor

const (
	Server         string = "http://192.168.1.13:27500"
	RegisterNode   string = "/monitor/nodeAdd"
	RegisterPod    string = "/monitor/podAdd"
	UnRegisterNode string = "/monitor/nodeDel/:hostname"
	UnRegisterPod  string = "/monitor/podDel/:namespace/:name"
)
