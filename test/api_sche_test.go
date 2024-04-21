package test

import (
	"minik8s/pkg/apiserver/app"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/scheduler"
	"testing"
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
	for i := 0; i < 10; i++ {
		apiServerDummy.MsgToScheduler()
	}
}
