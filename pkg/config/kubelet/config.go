package kubelet

const AddPod string = "/pod/AddPod"
const DelPod string = "/pod/DelPod/:namespace/:name"
const GetMatrix string = "/pod/GetMatrix/:namespace/:name"

const Cpu string = "2"
const Memory string = "8Gi"
const IpAddress string = "192.168.1.13"
const Port string = "20000"
