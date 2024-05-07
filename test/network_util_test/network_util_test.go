package test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/apiserver/app"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/network"
	"minik8s/tools"
)

var apiServerDummy *app.ApiServer = nil

func TestMain(m *testing.M) {
	tools.Apiserver_boot_finished = false
	tools.Test_finished = false
	tools.Test_enabled = true

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

	apiServerDummy.EtcdWrap.DelAll()

	dataStr, err := network.GetRequest(uri)
	fmt.Printf("Get request received: dataStr: %s, err: %v\n", dataStr, err)

	uri = config.API_server_prefix + config.API_add_endpoint
	ep := &api_obj.Endpoint{}
	ep.MetaData.Name = "test"
	ep.MetaData.NameSpace = "default"
	ep_str, _ := json.Marshal(ep)
	_, _ = network.PostRequest(uri, []byte(ep_str))

	uri = config.API_server_prefix + config.API_delete_endpoint_prefix + "default/test"
	dataStr, err = network.DelRequest(uri)
	tools.Test_finished = true

	if err != nil {
		t.Errorf("Failed to send DEL request.\n")
	} else {
		fmt.Println(dataStr)
	}

	for {
		if tools.Test_finished == false {
			time.Sleep(100 * time.Millisecond)
		} else {
			close(apiServerDummy.Producer.Sig)
			apiServerDummy.Producer.Producer.Close()
			break
		}
	}
}
