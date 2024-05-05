package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/apiserver/config"
	"minik8s/tools"
)

func (s *ApiServer) AddEndpoint(c *gin.Context) {
	fmt.Printf("[apiserver/AddEndpoint] Try to add an endpoint.\n")

	new_ep := &api_obj.Endpoint{}
	err := c.ShouldBind(new_ep)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/AddEndpoint] Failed to parse endpoint, " + err.Error(),
		})
		return
	}

	//存入etcd
	e_key := config.ETCD_endpoint_prefix + new_ep.MetaData.NameSpace + "/" + new_ep.MetaData.Name
	ep_str, err := json.Marshal(new_ep)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddEndpoint] Failed to marshal endpoint, " + err.Error(),
		})
		return
	}

	err = s.EtcdWrap.Put(e_key, ep_str)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddEndpoint] Failed to write into etcd, " + err.Error(),
		})
		return
	}

	if tools.Test_enabled {
		tools.Count_Test_Endpoint_Create += 1
	}

	//返回200
	c.JSON(http.StatusOK, gin.H{
		"data": "[handler/AddEndpoint] Add endpoint success",
	})
}

func (s *ApiServer) DeleteEndpoints(c *gin.Context) {
	fmt.Printf("[apiserver/DeleteEndpoints] Try to delete endpoints.\n")

	namespace := c.Param("namespace")
	srvname := c.Param("srvname")
	if srvname == "" || namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/DeleteEndpoints] Service name and namespace shall not be null.",
		})
		return
	}

	err := s.EtcdWrap.DeleteByPrefix(config.ETCD_endpoint_prefix + namespace + "/" + srvname)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/DeleteEndpoints] Failed to delete from etcd, " + err.Error(),
		})
		return
	}

	//返回200
	c.JSON(http.StatusOK, gin.H{
		"data": "[handler/DeleteEndpoints] Delete endpoints success",
	})
}

func (s *ApiServer) DeleteEndpoint(c *gin.Context) {
	fmt.Printf("[apiserver/DeleteEndpoint] Try to delete an endpoint.\n")

	namespace := c.Param("namespace")
	name := c.Param("name")
	if name == "" || namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/DeleteEndpoint] Service name and namespace shall not be null.",
		})
		return
	}

	fmt.Printf("[apiserver/DeleteEndpoint] namespace: %s, name: %s\n", name, namespace)

	err := s.EtcdWrap.Del(config.ETCD_endpoint_prefix + namespace + "/" + name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/DeleteEndpoint] Failed to delete from etcd, " + err.Error(),
		})
		return
	}

	//返回200
	c.JSON(http.StatusOK, gin.H{
		"data": "[handler/DeleteEndpoint] Delete endpoint success",
	})
}
