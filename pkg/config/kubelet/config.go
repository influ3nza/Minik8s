package kubelet

const Port int32 = 20000
const AddPod string = "/pod/AddPod"
const DelPod_prefix = "/pod/DelPod/"
const DelPod string = "/pod/DelPod/:namespace/:name/:pause"
const GetMetrics_prefix = "/pod/GetMetrics/"
const GetMetrics string = "/pod/GetMetrics/:namespace/:name"
const GetCpuAndMem string = "/pod/GetCpuAndMem"

const (
	MountNfs          string = "/pv/MountNfs/"
	UnmountNfs_prefix string = "/pv/UnmountNfs/"
	UnmountNfs        string = "/pv/UnmountNfs/:path"
)
