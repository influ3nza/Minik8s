package kubelet

const Port int32 = 20000
const AddPod string = "/pod/AddPod"
const DelPod_prefix string = "/pod/DelPod/"
const DelPod string = "/pod/DelPod/:namespace/:name/:pause"
const GetMatrix_prefix = "/pod/GetMatrix/"
const GetMatrix string = "/pod/GetMatrix/:namespace/:name"
