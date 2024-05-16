package tools

var Test_enabled bool = false
var Test_finished bool = false
var Apiserver_boot_finished bool = false
var Pod_created bool = false
var Ep_created bool = false
var Count_Test_Endpoint_Create int32 = 0
var NodesIpMap map[string]string = nil
var ClusterIpFlag int32 = 2
