package apiserver

const (
	//"registry/nodeip/:nodename/"
	ETCD_node_ip_prefix = "/registry/nodeip/"

	ETCD_node_prefix = "/registry/nodes/"

	//"/registry/pods/:namespace/:name"
	ETCD_pod_prefix = "/registry/pods/"

	//"/registry/services/:namespace/:name"
	ETCD_service_prefix = "/registry/services/"

	//"/registry/endpoints/:namespace/:name"
	//其中:name的格式为{service_name}-{pod_name}
	ETCD_endpoint_prefix = "/registry/endpoints/"

	ETCD_replicaset_prefix = "/registry/replicasets/"

	ETCD_hpa_prefix = "/registry/hpas/"

	ETCD_workflow_prefix = "/registry/workflows/"
)
