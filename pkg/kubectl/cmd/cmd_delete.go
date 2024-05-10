package cmd

import (
	"fmt"
	"strings"

	"minik8s/pkg/config/apiserver"

	"github.com/spf13/cobra"
)

var DelCmd = &cobra.Command{
	Use:     "delete [<apiobject> <namespace>/<name>]",
	Short:   "delete an apiobject",
	Long:    "This is a command to delete an apiobject. Simply use \"kubectl delete <apiobject> <namespace>/<name>\".",
	Example: "kubectl delete pod testing/pod1",
	Args:    cobra.ExactArgs(2),
	Run:     DelHandler,
}

func DelHandler(cmd *cobra.Command, args []string) {
	apitype := args[0]
	key := args[1]
	namespace, name := "", ""
	if strings.Count(key, "/") == 1 {
		index := strings.Index(key, "/") + 1
		namespace = key[0 : index-1]
		name = key[index:]

		if len(namespace) == 0 || len(name) == 0 {
			fmt.Println("[ERR] Empty namespace or name. Try -h for help.")
			return
		}
	} else {
		fmt.Println("[ERR] Wrong format. Try -h for help.")
		return
	}

	switch apitype {
	case "pod":
		DeletePodHandler(namespace, name)
	case "service":
		DeleteServiceHandler(namespace, name)
	case "replicaset":
		DeleteReplicasetHandler(namespace, name)
	case "hpa":
		DeleteHpaHandler(namespace, name)
	default:
		fmt.Println("[ERR] Wrong api kind. Available: pod, service, replicaset, hpa.")
	}
}

func DeletePodHandler(namespace string, name string) {
	uri := apiserver.API_server_prefix + apiserver.API_delete_pod_prefix + namespace + "/" + name
}

func DeleteServiceHandler(namespace string, name string) {

}

func DeleteReplicasetHandler(namespace string, name string) {

}

func DeleteHpaHandler(namespace string, name string) {

}
