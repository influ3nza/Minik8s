package utils

import (
	"minik8s/pkg/api_obj"
)

func CompareLabels(a map[string]string, b map[string]string) bool {
	for key, val := range a {
		if b[key] != val {
			return false
		}
	}
	return true
}

func CreateEndpoint(srv api_obj.Service, pod api_obj.Pod) error {
	//TODO:
	return nil
}
