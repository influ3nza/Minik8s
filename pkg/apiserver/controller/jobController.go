package controller

import (
	"fmt"
)

type JobController struct {
}

func (jc *JobController) PrintHandlerWarning() {
	fmt.Printf("[WARN/ReplicasetController] Error in message handler, the system may not be working properly!\n")
}

func CreateJobControllerInstance() (*JobController, error) {
	return &JobController{}, nil
}
