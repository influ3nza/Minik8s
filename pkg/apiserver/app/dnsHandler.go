package app

import (
	"minik8s/pkg/api_obj"
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
	dns := &api_obj.Dns{}
	err := c.ShouldBind(dns)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/DeleteDns] Failed to parse dns from request, " + err.Error(),
		})
		return
	}

	dns_namespace := dns.MetaData.NameSpace
	dns_name := dns.MetaData.Name
	if dns_namespace == "" || dns_name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[msgHandler/DeleteDns] Empty dns name or namespace",
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