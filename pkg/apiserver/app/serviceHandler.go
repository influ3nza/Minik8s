package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/config/kube_proxy"
	"minik8s/pkg/network"
	"minik8s/tools"
)

func (s *ApiServer) GetServices(c *gin.Context) {
	fmt.Printf("[apiserver/GetServices] Try to get all services.\n")

	res, err := s.EtcdWrap.GetByPrefix(apiserver.ETCD_service_prefix)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[apiserver/GetServices] Failed to get services, " + err.Error(),
		})
		return
	} else {
		var srvs = "["
		for id, srv := range res {
			srvs += srv.Value

			//返回值以逗号隔开
			if id < len(res)-1 {
				srvs += ","
			}
		}

		srvs += "]"

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

	//分配IP（检查IP）
	new_service.Spec.ClusterIP = AllocateClusterIp()
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
	//需要拿到所有proxy的地址。
	for _, nodeIp := range tools.NodesIpMap {
		//需要对接路径。
		uri := nodeIp + strconv.Itoa(int(kube_proxy.Port)) + kube_proxy.CreateService
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

func (s *ApiServer) DeleteService(c *gin.Context) {
	fmt.Printf("[apiserver/DeleteService] Try to delete a service.\n")

	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[apiserver/DeleteService] Empty namespace or name.",
		})
		return
	}

	//查找etcd中有没有这个srv
	e_key := apiserver.ETCD_service_prefix + namespace + "/" + name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/DeleteService] Failed to get from etcd, " + err.Error(),
		})
		return
	}

	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/DeleteService] Found zero or more than one srv.",
		})
		return
	}

	//找到有之后，就直接删除。
	srv := &api_obj.Service{}
	err = json.Unmarshal([]byte(res[0].Value), srv)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/DeleteService] Failed to unmarshal data, " + err.Error(),
		})
		return
	}

	err = s.EtcdWrap.Del(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/DeleteService] Failed to delete from etcd, " + err.Error(),
		})
		return
	}

	//删除ep
	uri := apiserver.API_server_prefix + apiserver.API_delete_endpoints_prefix + srv.MetaData.NameSpace + "/" + srv.MetaData.Name
	_, err = network.DelRequest(uri)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/DeleteService] Failed to send DEL request, " + err.Error(),
		})
		return
	}

	//给proxy发送消息。
	for _, nodeIp := range tools.NodesIpMap {
		//需要对接路径。
		uri := nodeIp + strconv.Itoa(int(kube_proxy.Port)) + kube_proxy.DeleteService
		_, err := network.PostRequest(uri, []byte(res[0].Value))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[ERR/handler/DeleteService] Failed to send POST request, " + err.Error(),
			})
			return
		}
	}

	//返回200
	c.JSON(http.StatusOK, gin.H{
		"data": "[handler/DeleteService] Delete service success",
	})
}

func (s *ApiServer) GetService(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[apiserver/GetService] Empty namespace or name.",
		})
		return
	}

	e_key := apiserver.ETCD_service_prefix + namespace + "/" + name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[apiserver/GetService] Empty namespace or name.",
		})
		return
	}

	if len(res) != 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[apiserver/GetService] Found zero or more than one srv.",
		})
		return
	}

	// 返回200
	c.JSON(http.StatusOK, gin.H{
		"data": res[0].Value,
	})
}
