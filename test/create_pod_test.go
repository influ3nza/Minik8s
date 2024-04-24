package test

import (
	"minik8s/pkg/apiserver/app"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/scheduler"
	"minik8s/pkg/kubectl/api"
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
func TestCreatePod(t *testing.T) {
	//读取yaml文件
	api.ParsePod("filename")

	//TODO: new test logic
}
