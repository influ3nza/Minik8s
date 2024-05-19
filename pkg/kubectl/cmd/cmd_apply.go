package cmd

import (
	"encoding/json"
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/kubectl/api"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

var ApplyCmd = &cobra.Command{
	Use:     "apply",
	Short:   "kubectl apply your filename",
	Long:    "This is a command for user to apply any fixed style yaml/json file. Simply use \"kubectl apply -f <your config file>\".",
	Example: "kubectl apply -f file_name.yaml ",
	Run:     ApplyHandler,
}

func init() {
	ApplyCmd.PersistentFlags().StringSliceVarP(&ApplyFiles, "file", "f", []string{}, "put your config files")
	err := ApplyCmd.MarkPersistentFlagRequired("file")
	if err != nil {
		fmt.Println("[ERR] Init Apply Failed.", err.Error())
		return
	}
}

func ApplyHandler(cmd *cobra.Command, args []string) {
	if ApplyFiles == nil || len(ApplyFiles) == 0 {
		fmt.Println("[ERR] Input file is empty.")
		return
	}

	for _, file := range ApplyFiles {
		var readData map[string]interface{}
		var format string
		if strings.HasSuffix(file, ".json") {
			format = "json"
		} else if strings.HasSuffix(file, ".yaml") {
			format = "yaml"
		} else {
			fmt.Println("[ERR] File must be json or yaml.")
			return
		}

		var fileToJson []byte
		var err error = nil
		if format == "json" {
			fileToJson, err = os.ReadFile(file)
			if err != nil {
				fmt.Printf("[ERR] Cannot read file %s.", file)
				return
			}
		} else {
			fileToJson, err = os.ReadFile(file)
			if err != nil {
				fmt.Printf("[ERR] Cannot read file %s.", file)
				return
			}

			fileToJson, err = yaml.YAMLToJSON(fileToJson)
			if err != nil {
				fmt.Printf("[ERR] Cannot convert yaml to json %s.", file)
				return
			}
		}
		err = json.Unmarshal(fileToJson, &readData)
		if err != nil {
			fmt.Println("[ERR] Fail to unmarshal json bytes.")
			return
		}

		kindValue, found := readData["kind"]
		if !found {
			fmt.Println("[ERR] Fail to find \"kind\" Keyword.")
			return
		}

		kind, ok := kindValue.(string)
		if !ok {
			fmt.Println("[ERR] Failed to get \"kind\" str.")
			return
		}

		switch strings.ToLower(kind) {
		case "pod":
			{
				// var pod = &api_obj.Pod{}
				// err = json.Unmarshal(fileToJson, pod)
				// if err != nil {
				// 	fmt.Printf("[ERR] Cannot parse file to pod, err: %s\n", err.Error())
				// 	return
				// }
				// fmt.Print(*pod)

				// err = api.SendObjectTo(fileToJson, "pod")
				err = api.ParsePod(file)
				if err != nil {
					fmt.Printf("[ERR] Cannot send pod to server, err: %s\n", err.Error())
					return
				}
			}
		case "node":
			{
				var node = &api_obj.Node{}
				err = json.Unmarshal(fileToJson, node)
				if err != nil {
					fmt.Printf("[ERR] Cannot parse file to node, err: %s\n", err.Error())
					return
				}
				fmt.Print(*node)

				err = api.SendObjectTo(fileToJson, "node")
				if err != nil {
					fmt.Printf("[ERR] Cannot send node to server, err: %s\n", err.Error())
					return
				}
			}
		case "service":
			{
				var service = &api_obj.Service{}
				err = json.Unmarshal(fileToJson, service)
				if err != nil {
					fmt.Printf("[ERR] Cannot parse file to service, err: %s\n", err.Error())
					return
				}
				fmt.Print(*service)

				err = api.SendObjectTo(fileToJson, "service")
				if err != nil {
					fmt.Printf("[ERR] Cannot send service to server, err: %s\n", err.Error())
					return
				}
			}
		}
	}
}
