package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/message"
	"minik8s/tools"
	"minik8s/pkg/config/apiserver"
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

	//创建endpoints -> 向endpointController发送消息。
	msg := &message.Message{
		Type:    message.SRV_CREATE,
		Content: string(service_str),
	}

	s.Producer.Produce(message.TOPIC_EndpointController, msg)

	//返回200
	c.JSON(http.StatusOK, gin.H{
		"data": "[handler/AddService] Add service success",
	})

	//TODO:通知proxy，service已经创建
}
