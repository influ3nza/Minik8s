package app

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"minik8s/pkg/apiserver/config"
)

func (s *ApiServer) GetNodes(c *gin.Context) {
	fmt.Printf("[apiserver/GetNodes] Try to get all nodes")

	res, err := s.EtcdWrap.GetByPrefix(config.ETCD_node_prefix)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[apiserver/GetNodes] Faled to get nodes, " + err.Error(),
		})
		return
	} else {
		var nodes []string
		for _, node := range res {
			nodes = append(nodes, node.Value)
		}

		c.JSON(http.StatusOK, gin.H{
			"data": nodes,
		})
	}
}

func (s *ApiServer) GetNode(c *gin.Context) {
	name := c.Param("name")
	fmt.Printf("[apiserver/GetNode] Try to get node: %s", name)

	if name == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "[apiserver/GetNode] Empty name string",
		})
		return
	} else {
		res, err := s.EtcdWrap.Get(config.ETCD_node_prefix + name)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "[apiserver/GetNode] Failed to get node, " + err.Error(),
			})
			return
		}

		if len(res) == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "[apiserver/GetNode] Failed to find node",
			})
			return
		}
		if len(res) != 1 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[apiserver/GetNode] Found more than one node",
			})
			return
		}

		result := res[0].Value
		c.JSON(http.StatusOK, gin.H{
			"data": result,
		})
		return
	}
}
