package test

import (
	"fmt"
	"testing"
	"time"

	"minik8s/pkg/apiserver/app"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/network"
	"minik8s/tools"
)

var apiServerDummy *app.ApiServer = nil

func TestMain(m *testing.M) {
	tools.Apiserver_boot_finished = false
	tools.Test_finished = false

	var err error
	apiServerDummy, err = app.CreateApiServerInstance(config.DefaultServerConfig())
	if err != nil {
		_ = fmt.Errorf("[ERR/network_util_test] Failed to create apiserver instance\n")
		return
	}

	go apiServerDummy.Run()
	m.Run()
}

func TestSendReceive(t *testing.T) {
	tools.Test_finished = false
	uri := config.API_server_prefix + config.API_get_nodes

	for {
		if tools.Apiserver_boot_finished == false {
			time.Sleep(100 * time.Millisecond)
		} else {
			break
		}
	}

	dataStr, errStr, err := network.GetRequest(uri)
	fmt.Printf("Get request received: dataStr: %s, errStr: %s, err: %v\n", dataStr, errStr, err)
	tools.Test_finished = true

	for {
		if tools.Test_finished == false {
			time.Sleep(100 * time.Millisecond)
		} else {
			break
		}
	}
}
