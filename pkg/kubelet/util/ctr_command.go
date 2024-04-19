package util

import (
	"fmt"
	"os/exec"
)

var path, _ = exec.LookPath("nerdctl")

func Exec(namespace string, args ...string) (string, error) {
	str := []string{"-n", namespace}
	str = append(str, args...)
	res, err := exec.Command(path, str...).CombinedOutput()
	return string(res), err
}

func PrintCmd(namespace string, args ...string) string {
	str := []string{"-n", namespace}
	str = append(str, args...)
	fmt.Println(str)
	return str[0]
}
