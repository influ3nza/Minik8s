package util

import (
	"fmt"
	"testing"
)

func TestPrintCmd(t *testing.T) {
	res := PrintCmd("test", "pull", "redis:latest")
	fmt.Println(res)
}
