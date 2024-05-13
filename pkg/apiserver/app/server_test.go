package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"minik8s/pkg/config/apiserver"
)

var apiServerDummy *ApiServer = nil

func TestMain(m *testing.M) {
	server, err := CreateApiServerInstance(apiserver.DefaultServerConfig())
	if err != nil {
		_ = fmt.Errorf("[ERR/server_test/main] Failed to create apiserver instance.\n")
		return
	}

	apiServerDummy = server
	go apiServerDummy.Run()
	m.Run()
}

func TestGet(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:50000/hello", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	apiServerDummy.router.ServeHTTP(w, req)
	fmt.Println(w.Body.String())

	// 注意：直接在本地运行此测试可能会被防火墙拦截。
}
