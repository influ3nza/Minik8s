package main

import "fmt"
import "SE3356/pkg/kubelet/util"

func main() {
	res := util.PrintCmd("test", "pull", "redis:latest")
	fmt.Println(res)
}
