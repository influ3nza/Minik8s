package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var hello_cmd = &cobra.Command{
	Use:   "hello",
	Short: "Hello kubectl!",
	Long:  "Hello kubectl! Put any args after this to test.",
	Run:   helloCmd_handler,
}

func helloCmd_handler(cmd *cobra.Command, args []string) {
	args_cnt := len(args)
	fmt.Printf("Hello kubectl! We get %d args, which is/are: %v\n", args_cnt, args)
}
