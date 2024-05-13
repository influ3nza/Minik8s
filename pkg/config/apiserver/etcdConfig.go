package apiserver

const (
	ETCD_node_prefix = "/registry/nodes/"

	//"/registry/pods/:namespace/:name"
	ETCD_pod_prefix = "/registry/pods/"

	//"/registry/services/:namespace/:name"
	ETCD_service_prefix = "/registry/services"

	//"/registry/endpoints/:namespace/:podUUID"
)
