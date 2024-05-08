package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var GetCmd = &cobra.Command{
	Use:     "get [<apiobject> <namespace>/<name>]",
	Short:   "gets the information of an apiobject",
	Long:    "This is a command for user to check the information of an apiobject. Simply use \"kubectl get <apiobject> <namespace>/<name>\".",
	Example: "kubectl get pod testing/pod1",
	Args:    cobra.ExactArgs(2),
	Run:     GetHandler,
}

func GetHandler(cmd *cobra.Command, args []string) {
	apitype := args[0]
	key := args[1]
	if strings.Count(key, "/") != 1 {
		fmt.Println("[ERR] ApiObj key in wrong format. Try -h for help.")
		return
	}

	index := strings.Index(key, "/") + 1
	namespace := key[0 : index-1]
	name := key[index:]

	if len(namespace) == 0 || len(name) == 0 {
		fmt.Println("[ERR] Empty namespace or name. Try -h for help.")
		return
	}

	switch apitype {
	case "pod":
	case "node":
	case "service":
	case "replicaset":
	case "hpa":
	default:
		fmt.Println("[ERR] Wrong api kind. Available: pod, node, service, replicaset, hpa.")
		return
	}
}
