package config

const (
	ETCD_node_prefix = "/registry/nodes/"

	//"/registry/pods/:namespace/:name"
	ETCD_pod_prefix = "/registry/pods/"

	//"/registry/services/:namespace/:name"
	ETCD_service_prefix = "/registry/services"

	//"/registry/endpoints/:namespace/:name"
	//其中:name的格式为{service_name}-{pod_name}
	ETCD_endpoint_prefix = "/registry/endpoints"
)
