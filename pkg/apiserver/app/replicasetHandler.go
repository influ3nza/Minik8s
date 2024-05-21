package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/config/apiserver"
)

func (s *ApiServer) AddReplicaSet(c *gin.Context) {
	fmt.Printf("[replicasethandler/AddReplicaSet] Try to add an ReplicaSet.\n")

	newReplicaSet := &api_obj.ReplicaSet{}
	err := c.ShouldBind(newReplicaSet)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/replicasethandler/AddReplicaSet] Failed to parse ReplicaSet, " + err.Error(),
		})
		return
	}

	//存入etcd
	r_key := apiserver.ETCD_replicaset_prefix + newReplicaSet.MetaData.NameSpace + "/" + newReplicaSet.MetaData.Name
	rs_str, err := json.Marshal(newReplicaSet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/replicasethandler/AddReplicaSet] Failed to marshal replicaset, " + err.Error(),
		})
		return
	}

	err = s.EtcdWrap.Put(r_key, rs_str)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/replicasethandler/AddReplicaSet] Failed to write into etcd, " + err.Error(),
		})
		return
	}

	/*
		if tools.Test_enabled {
			tools.Count_Test_Endpoint_Create += 1
		}
	*/

	//返回200
	c.JSON(http.StatusOK, gin.H{
		"data": "[replicasethandler/AddReplicaSet] Add replicaset success",
	})
}

func (s *ApiServer) DeleteReplicaSet(c *gin.Context) {
	fmt.Printf("[apiserver/DeleteReplicaSet] Try to delete replicaSet.\n")

	namespace := c.Param("namespace")
	name := c.Param("name")
	if name == "" || namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/replicasethandler/DeleteReplicaSet] Service name and namespace shall not be null.",
		})
		return
	}
	fmt.Printf("[replicasethandler/DeleteReplicaSet] namespace: %s, name: %s\n", name, namespace)

	err := s.EtcdWrap.Del(apiserver.ETCD_replicaset_prefix + namespace + "/" + name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/replicasethandler/DeleteReplicaSet] Failed to delete from etcd, " + err.Error(),
		})
		return
	}

	//返回200
	c.JSON(http.StatusOK, gin.H{
		"data": "[replicasethandler/DeleteReplicaSet] Delete replicaset success",
	})
}

func (s *ApiServer) GetReplicaSets(c *gin.Context) {
	fmt.Printf("[replicasethandler/GetReplicaSets] Try to get ReplicaSets.\n")
	key := apiserver.ETCD_replicaset_prefix
	replicasets, err := s.EtcdWrap.GetByPrefix(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/replicasethandler/GetReplicaSets] Failed to get from etcd, " + err.Error(),
		})
		return
	}

	var rss []string
	for id, ep := range replicasets {
		rss = append(rss, ep.Value)

		//返回值以逗号隔开
		if id < len(replicasets)-1 {
			rss = append(rss, ",")
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": rss,
	})
}
