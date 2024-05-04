package message

const (
	SRV_CREATE string = "srv_create"
	SRV_DELETE string = "srv_delete"
	POD_CREATE string = "pod_create"
	POD_UPDATE string = "pod_update"
	POD_DELETE string = "pod_delete"
)

const (
	TOPIC_EndpointController string = "EndpointController"
	TOPIC_ApiServer_FromNode string = "ApiServerFromNode"
)

type Message struct {
	Type    string
	Content string
	Backup  string
	Backup2 string
}
