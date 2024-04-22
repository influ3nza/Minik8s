package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"minik8s/pkg/api_obj"
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

func (s *ApiServer) AddNode(c *gin.Context) {
	fmt.Printf("[apiserver/AddNode] Try to add a node")

	var node api_obj.Node

	//TODO: post请求 要加JSON请求标头
	err := c.ShouldBind(&node)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[apiserver/AddNode] Failed to parse node, " + err.Error(),
		})

		return
	}

	//检查node各项参数
	//name是否重复
	node_name := node.GetName()
	res, err := s.EtcdWrap.Get(config.ETCD_node_prefix + node_name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/addNode] Get node failed, " + err.Error(),
		})
		return
	}

	if len(res) != 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/addNode] Node name already exists, " + err.Error(),
		})
		return
	}

	//初始化
	node.NodeMetadata.UUID = "default UUID"
	node.NodeStatus = api_obj.NodeStatus{}

	//parse， 此时的node已经是结构体了
	node_json, err := json.Marshal(node)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/addNode] Failed to marshal data, " + err.Error(),
		})
		return
	}

	err = s.EtcdWrap.Put(config.ETCD_node_prefix+node_name, node_json)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/addNode] Failed to write in etcd, " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "[msgHandler/addNode] Add node success",
	})
}
