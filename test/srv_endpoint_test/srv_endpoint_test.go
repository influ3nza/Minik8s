package test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"minik8s/pkg/apiserver/app"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/controller"
	"minik8s/pkg/kubectl/api"
	"minik8s/tools"
)

var apiServerDummy *app.ApiServer = nil
var epControllerDummy *controller.EndpointController = nil

func TestMain(m *testing.M) {
	tools.Apiserver_boot_finished = false
	tools.Count_Test_Endpoint_Create = 0
	tools.Test_enabled = true

	var err error
	apiServerDummy, err = app.CreateApiServerInstance(config.DefaultServerConfig())
	if err != nil {
		_ = fmt.Errorf("[ERR/srv_endpoint_test] Failed to create apiserver instance\n")
		return
	}

	epControllerDummy, err = controller.CreateEndpointControllerInstance()
	if err != nil {
		_ = fmt.Errorf("[ERR/srv_endpoint_test] Failed to create ep controller instance\n")
		return
	}

	go apiServerDummy.Run()
	go epControllerDummy.Run()
	m.Run()
}

func TestSrvAndEndpoint(t *testing.T) {
	//清除所有记录
	apiServerDummy.EtcdWrap.DelAll()

	for i := 1; i < 3; i++ {
		err := api.ParsePod("../pkg/etcd/testfile/Pod-" + strconv.Itoa(i) + ".yaml")
		if err != nil {
			tools.Test_finished = true
			t.Errorf("[ERR/srv_endpoint_test] Test failed.\n")
		}
	}

	time.Sleep(1000 * time.Millisecond)

	err := api.ParseSrv("../pkg/etcd/testfile/Service-1.yaml")
	if err != nil {
		tools.Test_finished = true
		t.Errorf("[ERR/srv_endpoint_test] Test failed.\n")
	}

	for {
		if tools.Count_Test_Endpoint_Create < 2 {
			time.Sleep(100 * time.Millisecond)
		} else {
			close(apiServerDummy.Producer.Sig)
			apiServerDummy.Producer.Producer.Close()
			close(epControllerDummy.Consumer.Sig)
			epControllerDummy.Consumer.Consumer.Close()
			break
			// time.Sleep(100 * time.Millisecond)
		}
	}
}
