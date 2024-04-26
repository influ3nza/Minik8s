package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/kubectl/api"
	"os"
	"strings"
)

var ApplyCmd = &cobra.Command{
	Use:     "apply",
	Short:   "kubectl apply your filename",
	Long:    "This is a command for user to apply any fixed style yaml/json file. Simply use \"kubectl apply -f <your config file>\".",
	Example: "kubectl apply -f IDOLOVESE3356.yaml ",
	Run:     ApplyHandler,
}

func init() {
	ApplyCmd.PersistentFlags().StringSliceVarP(&ApplyFiles, "file", "f", []string{}, "put your config files")
	err := ApplyCmd.MarkFlagRequired("file")
	if err != nil {
		fmt.Println("init Apply Failed at line 20", err.Error())
		return
	}
}

func ApplyHandler(cmd *cobra.Command, args []string) {
	if ApplyFiles == nil || len(ApplyFiles) == 0 {
		fmt.Println("Error! Input Files is Empty")
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
			fmt.Println("Error! File must be json or yaml")
			return
		}

		var fileToJson []byte
		var err error = nil
		if format == "json" {
			fileToJson, err = os.ReadFile(file)
			if err != nil {
				fmt.Printf("Error! Cannot read file %s", file)
				return
			}
		} else {
			fileToJson, err = os.ReadFile(file)
			if err != nil {
				fmt.Printf("Error! Cannot read file %s", file)
				return
			}

			fileToJson, err = yaml.YAMLToJSON(fileToJson)
			if err != nil {
				fmt.Printf("Error! Cannot convert yaml to json %s", file)
				return
			}
		}
		err = json.Unmarshal(fileToJson, &readData)
		if err != nil {
			fmt.Println("Error! Fault in unmarshal json bytes")
			return
		}

		kindValue, found := readData["kind"]
		if !found {
			fmt.Println("Error! Fail to find \"kind\" KeyWord")
			return
		}

		kind, ok := kindValue.(string)
		if !ok {
			fmt.Println("Error! Get \"kind\" str failed")
			return
		}

		switch strings.ToLower(kind) {
		case "pod":
			{
				var pod = &api_obj.Pod{}
				err = json.Unmarshal(fileToJson, pod)
				if err != nil {
					fmt.Printf("Error! Cannot parse file to pod, err: %s", err.Error())
					return
				}
				fmt.Print(*pod)

				err = api.SendPodTo(fileToJson)
				if err != nil {
					fmt.Printf("Error! Cannot send pod to server, err: %s\n", err.Error())
					return
				}
			}
		case "node":
			{
				var node = &api_obj.Node{}
				err = json.Unmarshal(fileToJson, node)
				if err != nil {
					fmt.Printf("Error! Cannot parse file to pod, err: %s", err.Error())
					return
				}
				fmt.Print(*node)

				err = api.SendNodeTo(fileToJson)
				if err != nil {
					fmt.Printf("Error! Cannot send pod to server, err: %s\n", err.Error())
					return
				}
			}
		}
	}
}
