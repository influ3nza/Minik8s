package message

const (
	ENDPOINT_SRV_CREATE string = "endpoint_srv_create"
	ENDPOINT_SRV_DELETE string = "endpoint_srv_delete"
	ENDPOINT_POD_UPDATE string = "endpoint_pod_update"
	ENDPOINT_POD_DELETE string = "endpoint_pod_delete"
)

const (
	TOPIC_EndpointController string = "EndpointController"
)

type Message struct {
	Type    string
	Content string
	Backup  string
	Backup2 string
}
