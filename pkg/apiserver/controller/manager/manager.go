package manager

import (
	"fmt"
	"minik8s/pkg/apiserver/controller"
	"os"
	"os/signal"
	"syscall"
)

type ControllerManager struct {
	EpController  *controller.EndpointController
	RsController  *controller.ReplicasetController
	HpaController *controller.HPAController
}

func CreateNewControllerManagerInstance() (ControllerManager, error) {
	ep, err := controller.CreateEndpointControllerInstance()
	if err != nil {
		fmt.Printf("[Controller/MAIN] Failed to create ep controller.")
		return ControllerManager{}, err
	}

	rs, err := controller.CreateReplicasetControllerInstance()
	if err != nil {
		fmt.Printf("[Controller/MAIN] Failed to create ep controller.")
		return ControllerManager{}, err
	}

	return ControllerManager{
		EpController: ep,
		RsController: rs,
	}, nil
}

func (cm *ControllerManager) Run() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	go func() {
		<-sigChan
		cm.Clean()
	}()

	go cm.EpController.Run()
	go cm.RsController.Run()
}

func (cm *ControllerManager) Clean() {
	fmt.Printf("[Controller/CLEAN] Controller closing...\n")

	close(cm.EpController.Consumer.Sig)
	cm.EpController.Consumer.Consumer.Close()
	os.Exit(0)
}
