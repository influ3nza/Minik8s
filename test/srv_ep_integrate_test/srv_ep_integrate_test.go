package test

import (
	"fmt"
	"testing"
	"time"

	"minik8s/pkg/apiserver/app"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/controller"
	"minik8s/pkg/kubectl/api"
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

func TestSrvCreate(t *testing.T) {
	tools.Test_finished = false
	tools.Pod_created = false

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
		if tools.Pod_created == false {
			time.Sleep(10 * time.Millisecond)
		} else {
			tools.Pod_created = false
			break
		}
	}

	err = api.ParseSrv("../yamls/Srv-1.yaml")
	if err != nil {
		tools.Test_finished = true
		t.Errorf("[ERR/connection_test] Test failed.\n")
		return
	}

	for {
		if tools.Test_finished == false {
			time.Sleep(10 * time.Millisecond)
		} else {
			break
		}
	}
}
