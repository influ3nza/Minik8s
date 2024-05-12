package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/network"
	"minik8s/tools"
)

func (s *ApiServer) GetServices(c *gin.Context) {
	fmt.Printf("[apiserver/GetServices] Try to add all services.\n")

	res, err := s.EtcdWrap.GetByPrefix(apiserver.ETCD_service_prefix)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[apiserver/GetServices] Failed to get services, " + err.Error(),
		})
		return
	} else {
		var srvs []string
		for id, srv := range res {
			srvs = append(srvs, srv.Value)

			//返回值以逗号隔开
			if id < len(res)-1 {
				srvs = append(srvs, ",")
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"data": srvs,
		})
	}
}

func (s *ApiServer) AddService(c *gin.Context) {
	fmt.Printf("[apiserver/AddService] Try to add a service.\n")

	new_service := &api_obj.Service{}
	err := c.ShouldBind(new_service)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/AddService] Failed to parse service, " + err.Error(),
		})
		return
	}

	service_name := new_service.MetaData.Name
	service_namespace := new_service.MetaData.NameSpace

	if service_name == "" || service_namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/AddService] Empty service name or namespace",
		})
		return
	}

	fmt.Printf("[apiserver/AddService] Service name: %s\n", service_name)

	e_key := apiserver.ETCD_service_prefix + service_namespace + "/" + service_name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddService] Get service failed, " + err.Error(),
		})
		return
	}

	if len(res) != 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddService] Service name already exists.",
		})
		return
	}

	//TODO:分配IP（检查IP）
	new_service.Spec.ClusterIP = "10.0.0.12"
	new_service.MetaData.UUID = tools.NewUUID()
	new_service.Status = api_obj.ServiceStatus{
		Condition: api_obj.SERVICE_PENDING,
	}

	//存入etcd
	service_str, err := json.Marshal(new_service)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddService] Failed to marshal service, " + err.Error(),
		})
		return
	}

	err = s.EtcdWrap.Put(e_key, service_str)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddService] Failed to write into etcd, " + err.Error(),
		})
		return
	}

	//这里先不给endpoint controller发送srv create的消息。因为我们要保证
	//ep一定要创建在srv之后。

	//通知proxy，service已经创建，这里使用http过去，消息队列回来的方式。
	//循环遍历所有node上的proxy：
	//TODO:需要拿到所有proxy的地址。
	proxyipdummy := []string{}
	for _, proxyAddr := range proxyipdummy {
		//TODO: 需要对接路径。
		uri := proxyAddr + "/service/AddService"
		_, err := network.PostRequest(uri, service_str)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[ERR/handler/AddService] Failed to send POST request, " + err.Error(),
			})
			return
		}
	}

	//返回200
	c.JSON(http.StatusOK, gin.H{
		"data": "[handler/AddService] Add service success",
	})
}
