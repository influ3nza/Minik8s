package test

import (
	"encoding/json"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/apiserver/app"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/scheduler"
	"testing"
	"time"
)

var apiServerDummy *app.ApiServer = nil
var schedulerDummy *scheduler.Scheduler = nil

func TestMain(m *testing.M) {
	apiServerDummy, _ = app.CreateApiServerInstance(config.DefaultServerConfig())
	schedulerDummy, _ = scheduler.CreateSchedulerInstance()
	go apiServerDummy.Run()
	go schedulerDummy.Run()
	m.Run()
}

// 测试apiserver向scheduler发送消息
func TestSendMsg(t *testing.T) {
	node1 := &api_obj.NodeDummy{
		UUID: "ssss",
		Val:  "eeee",
	}
	c_node1, _ := json.Marshal(node1)

	apiServerDummy.EtcdWrap.DelAll()
	apiServerDummy.EtcdWrap.Put(config.ETCD_node_prefix+node1.UUID, c_node1)

	time.Sleep(1 * time.Second)

	for i := 0; i < 1; i++ {
		apiServerDummy.MsgToScheduler()
	}

	for {
		time.Sleep(1 * time.Second)
	}
}
