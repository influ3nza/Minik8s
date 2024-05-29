package tools

var Test_enabled bool = false
var Test_finished bool = false
var Apiserver_boot_finished bool = false
var Pod_created bool = false
var Ep_created bool = false
var Count_Test_Endpoint_Create int32 = 0
var NodesIpMap map[string]string = nil
var ClusterIpFlag int32 = 2
var PV_mount_master_path string = "/mnt/minik8s"
var PV_mount_node_path string = "/mnt/m_minik8s"
var PV_mount_params string = " 192.168.1.0/24(rw,sync,no_root_squash,no_subtree_check)"
var PV_mount_config_file string = "/etc/exports"
var Func_template_path string = "/ZTH/Minik8s/pkg/serverless/common/"