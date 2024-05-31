package cmd

import (
	"minik8s/pkg/kubectl/api"
	"fmt"

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
	err := api.ParseFunc(fileName)
	if err != nil {
		fmt.Printf("[ERR/UpdateFunction] Failed to update function.\n")
	}
}
