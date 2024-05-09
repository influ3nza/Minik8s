package message

const (
	POD_CREATE string = "pod_create"
	POD_UPDATE string = "pod_update"
	POD_DELETE string = "pod_delete"
)

const (
	SRV_CREATE string = "srv_create"
	SRV_DELETE string = "srv_delete"
	EP_ADD     string = "ep_add"
	EP_DELETE  string = "ep_delete"
)

const (
	TOPIC_EndpointController string = "EndpointController"
	TOPIC_ApiServer_FromNode string = "ApiServerFromNode"
)

const (
	DEL_POD_SUCCESS   string = "del_pod_success"
	DEL_POD_FAILED    string = "del_pod_failed"
	DEL_POD_NOT_EXIST string = "del_pod_not_exist"
)

type Message struct {
	Type    string
	Content string
	Backup  string
	Backup2 string
}
