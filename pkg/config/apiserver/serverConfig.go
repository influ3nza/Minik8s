package apiserver

import (
	"time"
)

const (
	API_default_namespace string = "default"
)

const (
	API_server_prefix string = "http://192.168.1.13:50000"

	API_get_nodes       string = "/nodes/getAll"
	API_get_node_prefix string = "/nodes/"
	API_get_node        string = "/nodes/:name"
	API_add_node        string = "/nodes/add"

	API_update_pod                    string = "/pods/update"
	API_add_pod                       string = "/pods/add"
	API_get_pod_prefix                string = "/pods/"
	API_get_pod                       string = "/pods/:namespace/:name"
	API_get_pods                      string = "/pods/getAll"
	API_get_pods_by_node_prefix       string = "/pods/getByNode/"
	API_get_pods_by_node              string = "/pods/getByNode/:nodename"
	API_get_pods_by_node_force_prefix string = "/pods/getByNodeForce/"
	API_get_pods_by_node_force        string = "/pods/getByNodeForce/:nodename"
	API_get_pods_by_namespace_prefix  string = "/pods/getByNamespace/"
	API_get_pods_by_namespace         string = "/pods/getByNamespace/:namespace"
	API_delete_pod_prefix             string = "/pods/delete/"
	API_delete_pod                    string = "/pods/delete/:namespace/:name"
	API_get_pod_metrics               string = "/pods/getMetrics/:namespace/:name"
	API_get_pod_metrics_prefix        string = "/pods/getMetrics/"

	API_add_service           string = "/services/add"
	API_get_services          string = "/services/getAll"
	API_get_service_prefix    string = "/services/"
	API_get_service           string = "/services/:namespace/:name"
	API_delete_service_prefix string = "/services/delete/"
	API_delete_service        string = "/services/delete/:namespace/:name"

	API_add_endpoint string = "/endpoints/add"
	//所有endpoint的名字{srvname}-{podname}
	API_delete_endpoints_prefix        string = "/endpoints/deleteBatch/"
	API_delete_endpoints               string = "/endpoints/deleteBatch/:namespace/:srvname"
	API_delete_endpoint_prefix         string = "/endpoints/delete/"
	API_delete_endpoint                string = "/endpoints/delete/:namespace/:name"
	API_get_endpoint_prefix            string = "/endpoints/"
	API_get_endpoint                   string = "/endpoints/:namespace/:name"
	API_get_endpoint_by_service_prefix string = "/endpoints/getBySrv/"
	API_get_endpoint_by_service        string = "/endpoints/getBySrv/:namespace/:srvname"

	API_get_replicasets          string = "/replicasets/getAll"
	API_delete_replicaset_prefix string = "/replicasets/delete/"
	API_delete_replicaset        string = "/replicasets/delete/:namespace/:name"
	API_update_replicaset        string = "/replicasets/update"
	API_add_replicaset           string = "/replicasets/add"

	API_scale_replicaset_prefix string = "/replicasets/scaleup/"
	API_scale_replicaset        string = "/replicasets/scaleup/:name/:method"

	API_add_hpa           string = "/hpas/add"
	API_get_hpas          string = "/hpas/getAll"
	API_delete_hpa_prefix string = "/hpas/delete/"
	API_delete_hpa        string = "/hpas/delete/:namespace/:name"
	API_update_hpa        string = "/hpas/update"

	API_add_dns           string = "/dns/add"
	API_get_dns_prefix    string = "/dns/"
	API_get_dns           string = "/dns/:namespace/:name"
	API_delete_dns_prefix string = "/dns/delete/"
	API_delete_dns        string = "/dns/delete/:namespace/:name"
	API_get_all_dns       string = "/dns/getAll"

	API_get_workflow_prefix    string = "/workflows/"
	API_get_workflow           string = "/workflows/:name"
	API_add_workflow           string = "/workflows/add"
	API_update_workflow        string = "/workflows/update"
	API_delete_workflow_prefix string = "/workflows/delete/"
	API_delete_workflow        string = "/workflows/delete/:name"
	API_exec_workflow_prefix   string = "/workflows/exec/"
	API_exec_workflow          string = "/workflows/exec/:name"
	API_check_workflow         string = "/workflow/check"

	API_get_function_prefix     string = "/functions/"
	API_get_function            string = "/functions/:name"
	API_get_all_functions       string = "/functions/getAll"
	API_add_function            string = "/functions/add"
	API_delete_function_prefix  string = "/functions/delete/"
	API_delete_function         string = "/functions/delete/:name"
	API_exec_function_prefix    string = "/functions/exec/"
	API_exec_function           string = "/functions/exec/:name"
	API_find_function_ip_prefix string = "/function/findByIp/"
	API_find_function_ip        string = "/function/findByIp/:name"
	API_get_function_res_prefix string = "/function/getRes/"
	API_get_function_res        string = "/function/getRes/:name"
	API_update_function         string = "/function/update"

	API_add_pv           string = "/pv/add"
	API_delete_pv_prefix string = "/pv/delete/"
	API_delete_pv        string = "/pv/delete/:name"
	API_get_pvs          string = "/pv/getAll"

	API_add_pvc           string = "/pvc/add"
	API_delete_pvc_prefix string = "/pvc/delete/"
	API_delete_pvc        string = "/pvc/delete/:name"
	API_get_pvcs          string = "/pvc/getAll"

	API_delete_registry string = "/deleteRegistry"
)

type ServerConfig struct {
	Port          int32
	TrustedProxy  []string
	EtcdEndpoints []string
	EtcdTimeout   time.Duration
	MaxNodeCount  int32
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Port:          50000,
		TrustedProxy:  []string{"127.0.0.1"},
		EtcdEndpoints: []string{"localhost:2379"},
		EtcdTimeout:   5 * time.Second,
		MaxNodeCount:  10,
	}
}
