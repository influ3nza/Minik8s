package kubelet

const Port int32 = 20000
const AddPod string = "/pod/AddPod"
const DelPod string = "/pod/DelPod/:namespace/:name"
const GetMatrix string = "/pod/GetMatrix/:namespace/:name"
