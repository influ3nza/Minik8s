package test

import (
	"fmt"
	"testing"
	"time"

	"minik8s/pkg/apiserver/app"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/controller"
	"minik8s/pkg/kubectl/api"
	"minik8s/pkg/network"
	"minik8s/pkg/scheduler"
	"minik8s/tools"
)

var apiServerDummy *app.ApiServer = nil
var schedulerDummy *scheduler.Scheduler = nil
var endpointCtrlDummy *controller.EndpointController = nil

func TestMain(m *testing.M) {
	tools.Apiserver_boot_finished = false
	tools.Test_finished = false
	tools.Test_enabled = true

	var err error
	apiServerDummy, err = app.CreateApiServerInstance(apiserver.DefaultServerConfig())
	if err != nil {
		_ = fmt.Errorf("Failed to create instance!")
	}
	schedulerDummy, err = scheduler.CreateSchedulerInstance()
	if err != nil {
		_ = fmt.Errorf("Failed to create instance!")
	}
	endpointCtrlDummy, err = controller.CreateEndpointControllerInstance()
	if err != nil {
		_ = fmt.Errorf("Failed to create instance!")
	}
	go apiServerDummy.Run()
	go schedulerDummy.Run()
	go endpointCtrlDummy.Run()
	m.Run()
}

func TestConnection(t *testing.T) {
	tools.Test_finished = false
	for {
		if tools.Apiserver_boot_finished == false {
			time.Sleep(10 * time.Millisecond)
		} else {
			break
		}
	}

	//清除所有记录
	apiServerDummy.EtcdWrap.DeleteByPrefix("/registry")

	//读取yaml文件
	err := api.ParseNode("./Node-1.yaml")
	if err != nil {
		tools.Test_finished = true
		t.Errorf("[ERR/connection_test] Test failed.\n")
		return
	}

	err = api.ParsePod("./Pod-1.yaml")
	if err != nil {
		tools.Test_finished = true
		t.Errorf("[ERR/connection_test] Test failed.\n")
		return
	}

	for {
		if tools.Test_finished == false {
			time.Sleep(10 * time.Millisecond)
		} else {
			tools.Test_finished = false
			break
		}
	}

	_, _ = network.DelRequest(apiserver.API_server_prefix + apiserver.API_delete_pod_prefix + "defaulttest/pod-example1")

	for {
		if tools.Test_finished == false {
			time.Sleep(10 * time.Millisecond)
		} else {
			close(schedulerDummy.Consumer.Sig)
			close(schedulerDummy.Producer.Sig)
			close(apiServerDummy.Producer.Sig)
			close(endpointCtrlDummy.Consumer.Sig)
			schedulerDummy.Consumer.Consumer.Close()
			schedulerDummy.Producer.Producer.Close()
			apiServerDummy.Producer.Producer.Close()
			endpointCtrlDummy.Consumer.Consumer.Close()
			break
		}
	}
}
