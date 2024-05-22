package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/config/apiserver"
	"minik8s/tools"
)

func (s *ApiServer) GetNodes(c *gin.Context) {
	fmt.Printf("[apiserver/GetNodes] Try to get all nodes.\n")

	res, err := s.EtcdWrap.GetByPrefix(apiserver.ETCD_node_prefix)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[apiserver/GetNodes] Failed to get nodes, " + err.Error(),
		})
		return
	} else {
		var nodes = "["
		for id, node := range res {
			nodes += node.Value

			//返回值以逗号隔开
			if id < len(res)-1 {
				nodes += ","
			}
		}

		nodes += "]"

		c.JSON(http.StatusOK, gin.H{
			"data": nodes,
		})
	}
}

func (s *ApiServer) GetNode(c *gin.Context) {
	name := c.Param("name")
	fmt.Printf("[apiserver/GetNode] Try to get node: %s.\n", name)

	if name == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "[apiserver/GetNode] Empty name string",
		})
		return
	} else {
		res, err := s.EtcdWrap.Get(apiserver.ETCD_node_prefix + name)
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

// WARN:这个函数仅供测试使用，在对接时需要进行修改。
func (s *ApiServer) AddNode(c *gin.Context) {
	fmt.Printf("[apiserver/addNode] Try to add a node.\n")

	var node api_obj.Node

	err := c.ShouldBind(&node)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/apiserver/addNode] Failed to parse node, " + err.Error(),
		})

		return
	}

	//检查node各项参数
	//name是否重复，是否为空
	node_name := node.GetName()
	fmt.Printf("[msgHandler/addNode] Adding node: %s.\n", node_name)

	if node_name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[msgHandler/addNode] Empty node name.",
		})
		return
	}

	fmt.Printf("[apiserver/addNode] Node name: %s\n", node_name)

	res, err := s.EtcdWrap.Get(apiserver.ETCD_node_prefix + node_name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/addNode] Get node failed, " + err.Error(),
		})
		return
	}

	if len(res) != 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/addNode] Node name already exists.",
		})
		return
	}

	//初始化
	node.NodeMetadata.UUID = tools.NewUUID()
	//这里是便于测试，之后需要重新书写
	node.NodeStatus.Condition = api_obj.Ready

	//存储node的ip地址
	nodeaddr := "http://" + node.NodeStatus.Addresses.InternalIp + ":"
	tools.NodesIpMap[node_name] = nodeaddr

	fmt.Printf("[msgHandler/addNode] Node internal Ip: %s.\n", nodeaddr)

	e_key := apiserver.ETCD_node_ip_prefix + node.NodeMetadata.Name
	err = s.EtcdWrap.Put(e_key, []byte(nodeaddr))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/AddNode] Failed to write in etcd, " + err.Error(),
		})
		return
	}

	//parse， 此时的node已经是结构体了
	node_json, err := json.Marshal(node)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/addNode] Failed to marshal data, " + err.Error(),
		})
		return
	}

	err = s.EtcdWrap.Put(apiserver.ETCD_node_prefix+node_name, node_json)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[msgHandler/addNode] Failed to write in etcd, " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": "[msgHandler/addNode] Add node success",
	})
}
