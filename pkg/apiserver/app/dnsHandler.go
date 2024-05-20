package app

import (
	"encoding/json"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/config/apiserver"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *ApiServer) AddDns(c *gin.Context) {
	dns := &api_obj.Dns{}
	err := c.ShouldBind(dns)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/AddDns] Failed to parse dns from request, " + err.Error(),
		})
		return
	}

	dns_namespace := dns.MetaData.NameSpace
	dns_name := dns.MetaData.Name
	if dns_namespace == "" || dns_name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[msgHandler/AddDns] Empty dns name or namespace",
		})
		return
	}

	//填写srv的clusterIp
	for _, path := range dns.Paths {
		e_key := apiserver.ETCD_service_prefix + dns_namespace + "/" + path.ServiceName
		res, err := s.EtcdWrap.Get(e_key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[msgHandler/AddDns] Failed to get from etcd, " + err.Error(),
			})
			return
		}
		if len(res) != 1 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[msgHandler/AddDns] Found zero or more than one srv.",
			})
			return
		}

		srv := &api_obj.Service{}
		err = json.Unmarshal([]byte(res[0].Value), srv)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[msgHandler/AddDns] Failed to unmarshal data, " + err.Error(),
			})
			return
		}

		path.ServiceIp = srv.Spec.ClusterIP
	}

	dns_str, err := json.Marshal(dns)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/AddDns] Failed to marshal data, " + err.Error(),
		})
		return
	}

	e_key := apiserver.ETCD_dns_prefix + dns_namespace + "/" + dns_name
	err = s.EtcdWrap.Put(e_key, dns_str)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddDns] Failed to put into etcd, " + err.Error(),
		})
		return
	}

	err = s.DnsService.AddDns(dns)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/AddDns] Failed to add dns, " + err.Error(),
		})
		return
	}

	//成功返回
	c.JSON(http.StatusCreated, gin.H{
		"data": "[msgHandler/AddDns] Create dns success",
	})
}

func (s *ApiServer) DeleteDns(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[msgHandler/DeleteDns] Empty dns name or namespace",
		})
		return
	}

	e_key := apiserver.ETCD_dns_prefix + namespace + "/" + name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/DeleteDns] Failed to get from etcd, " + err.Error(),
		})
		return
	}
	if len(res) != 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/DeleteDns] Found zero or more than one dns.",
		})
		return
	}

	dns := &api_obj.Dns{}
	err = json.Unmarshal([]byte(res[0].Value), dns)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/DeleteDns] Failed to unmarshal data, " + err.Error(),
		})
		return
	}

	err = s.DnsService.DeleteDns(dns)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/DeleteDns] Failed to add dns, " + err.Error(),
		})
		return
	}

	//成功返回
	c.JSON(http.StatusCreated, gin.H{
		"data": "[msgHandler/DeleteDns] Delete dns success",
	})
}
