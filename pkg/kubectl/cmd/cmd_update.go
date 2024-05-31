package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/kubectl/api"
	"minik8s/pkg/network"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

var UpdateCmd = &cobra.Command{
	Use:     "update",
	Short:   "kubectl update <filename>",
	Long:    "This is a command for user to update serverless functions only",
	Example: "kubectl update file_name.yaml ",
	Args:    cobra.ExactArgs(1),
	Run:     UpdateHandler,
}

func UpdateHandler(cmd *cobra.Command, args []string) {
	fileName := args[0]
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parseFunc] Failed to open file, err:%v\n", err)
		return
	}

	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parseFunc] Failed to read file, err:%v\n", err)
		return
	}
	f := &api_obj.Function{}

	err = yaml.Unmarshal(content, f)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parseFunc] Failed to unmarshal yaml, err:%v\n", err)
		return
	}

	fileToJson, _ := json.Marshal(f)
	UpdateFunctionHandler(fileToJson)
}

func UpdateFunctionHandler(fileToJson []byte) error {
	f := &api_obj.Function{}
	err := json.Unmarshal(fileToJson, f)
	if err != nil {
		return err
	}

	path := f.FilePath
	api.DoZip(path, path+".zip")
	content, err := os.ReadFile(path + ".zip")
	_ = os.Remove(path + ".zip")
	if err != nil {
		return err
	}

	fw := api_obj.FunctionWrap{
		Func:    *f,
		Content: content,
	}
	fw_str, err := json.Marshal(fw)
	if err != nil {
		return err
	}

	//发送请求。
	uri := apiserver.API_server_prefix + apiserver.API_update_function
	_, err = network.PostRequest(uri, fw_str)
	return err
}
