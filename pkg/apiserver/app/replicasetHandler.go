package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/apiserver/controller/utils"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/config/kubelet"
	"minik8s/pkg/network"
	"minik8s/tools"
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

	e_key := apiserver.ETCD_pod_prefix + namespace + "/" + utils.RS_prefix + name
	res, err := s.EtcdWrap.GetByPrefix(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/replicasethandler/DeleteReplicaSets] Failed to get from etcd, " + err.Error(),
		})
		return
	}

	//删除所有对应的pod
	for _, kv := range res {
		pod := &api_obj.Pod{}
		err := json.Unmarshal([]byte(kv.Value), pod)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[ERR/replicasethandler/DeleteReplicaSets] Failed to unmarshal data, " + err.Error(),
			})
			return
		}

		e_key := apiserver.ETCD_pod_prefix + pod.MetaData.NameSpace + "/" + pod.MetaData.Name
		err = s.EtcdWrap.Del(e_key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[ERR/replicasethandler/GetReplicaSets] Failed to delete from etcd, " + err.Error(),
			})
			return
		}

		uri := tools.NodesIpMap[pod.Spec.NodeName] + strconv.Itoa(int(kubelet.Port)) + kubelet.DelPod_prefix +
			pod.MetaData.NameSpace + "/" + pod.MetaData.Name + "/" + pod.MetaData.Annotations["pause"]
		_, err = network.DelRequest(uri)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[ERR/replicasethandler/GetReplicaSets] Failed to send DEL request, " + err.Error(),
			})
			return
		}
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

func (s *ApiServer) UpdateReplicaSet(c *gin.Context) {
	fmt.Printf("[apiserver/UpdateReplicaSet] Try to update replicaset.")

	rs := &api_obj.ReplicaSet{}
	err := c.ShouldBind(rs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/replicasethandler/UpdateReplicaSet] Failed to parse data into replicaset.",
		})
		return
	}

	e_key := apiserver.ETCD_replicaset_prefix + rs.MetaData.NameSpace + "/" + rs.MetaData.Name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/replicasethandler/UpdateReplicaSets] Failed to get from etcd, " + err.Error(),
		})
		return
	}
	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/replicasethandler/UpdateReplicaSets] Found zero or more than one rs.",
		})
		return
	}

	new_rs_str, err := json.Marshal(rs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/replicasethandler/UpdateReplicaSets] Failed to marshal data, " + err.Error(),
		})
		return
	}

	err = s.EtcdWrap.Put(e_key, new_rs_str)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/replicasethandler/UpdateReplicaSets] Failed to write into etcd, " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": "[replicasethandler/UpdateReplicaSets] Update rs success",
	})
}

func (s *ApiServer) ScaleReplicaSet(c *gin.Context) {
	name := c.Param("name")
	method := c.Param("method")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/replicasethandler/ScaleUpReplicaSet] Empty replicaset name.",
		})
		return
	}

	offset := 1
	if method != "add" {
		offset = -1
	}
	err := s.U_ScaleReplicaSet(name, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/replicasethandler/ScaleUpReplicaSet] Failed, " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": "[replicasethandler/ScaleUpReplicaSet] Scale rs success",
	})
}
