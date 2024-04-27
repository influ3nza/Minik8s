package cmd

import (
	"github.com/spf13/cobra"
)

var root_kubectl = &cobra.Command{
	Use:   "kubectl",
	Short: "Using kubectl to work with kubernetes.",
	Long:  "Using kubectl to work with kubernetes. For specific usage, see 'kubectl [command] --help'.",
}

func Execute() error {
	return root_kubectl.Execute()
}

func init() {
	root_kubectl.AddCommand(hello_cmd)
	// root_kubectl.AddCommand(ApplyCmd)
}

var ApplyFiles []string
