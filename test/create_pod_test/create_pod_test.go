package test

import (
	"minik8s/pkg/apiserver/app"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/kubectl/api"
	"minik8s/pkg/scheduler"
	"minik8s/tools"

	"strconv"
	"testing"
	"time"
)

var apiServerDummy *app.ApiServer = nil
var schedulerDummy *scheduler.Scheduler = nil

func TestMain(m *testing.M) {
	tools.Apiserver_boot_finished = false
	tools.Test_finished = false
	tools.Test_enabled = true

	apiServerDummy, _ = app.CreateApiServerInstance(config.DefaultServerConfig())
	schedulerDummy, _ = scheduler.CreateSchedulerInstance()
	go apiServerDummy.Run()
	go schedulerDummy.Run()
	m.Run()
}

// 测试apiserver向scheduler发送消息
func TestCreatePod(t *testing.T) {
	tools.Test_finished = false
	for {
		if tools.Apiserver_boot_finished == false {
			time.Sleep(100 * time.Millisecond)
		} else {
			break
		}
	}

	//清除所有记录
	apiServerDummy.EtcdWrap.DelAll()

	for i := 1; i < 3; i++ {
		err := api.ParseNode("Z:\\Minik8s\\pkg\\etcd\\testfile\\Node-" + strconv.Itoa(i) + ".yaml")
		if err != nil {
			tools.Test_finished = true
			t.Errorf("[ERR/create_pod_test] Test failed.\n")
		}
	}

	//读取yaml文件
	err := api.ParsePod("Z:\\Minik8s\\pkg\\etcd\\testfile\\Pod-1.yaml")
	if err != nil {
		tools.Test_finished = true
		t.Errorf("[ERR/create_pod_test] Test failed.\n")
	}

	for {
		if tools.Test_finished == false {
			time.Sleep(100 * time.Millisecond)
		} else {
			time.Sleep(2 * time.Second)
			// schedulerDummy.Consumer.Consumer.Close()
			// schedulerDummy.Producer.Producer.Close()
			// apiServerDummy.Producer.Producer.Close()
			break
		}
	}
}
