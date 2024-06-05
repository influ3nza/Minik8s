package app

import (
	"encoding/json"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/apiserver/controller/utils"
	"minik8s/pkg/config/apiserver"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *ApiServer) AddPVC(c *gin.Context) {
	//TODO
	pvc := &api_obj.PVC{}
	err := c.ShouldBind(pvc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/AddPVC] Failed to parse pvc, " + err.Error(),
		})
		return
	}

	name := pvc.Metadata.Name
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/AddPVC] Empty pvc name.",
		})
		return
	}

	//检查是否重复
	e_key := apiserver.ETCD_pvc_prefix + name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddPVC] Failed to get from etcd, " + err.Error(),
		})
		return
	}

	if len(res) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/AddPVC] PV name already exists.",
		})
		return
	}

	//查找并绑定PV
	p_key := apiserver.ETCD_pv_prefix
	p_res, err := s.EtcdWrap.GetByPrefix(p_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddPVC] Failed to get from etcd, " + err.Error(),
		})
		return
	}

	flag := false

	for _, kv := range p_res {
		pv := &api_obj.PV{}
		err := json.Unmarshal([]byte(kv.Value), pv)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[ERR/handler/AddPVC] Failed to unmarshal data, " + err.Error(),
			})
		}

		if utils.CompareLabels(pv.Metadata.Labels, pvc.Spec.Selector) {
			flag = true
			pvc.Spec.BindPV = pv.Metadata.Name
			break
		}
	}

	//动态创建PV
	if !flag {
		err = s.DynamicCreatePV(pvc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[ERR/handler/AddPVC] Failed to dynamically create pv, " + err.Error(),
			})
			return
		}
	}

	//写入etcd
	pvc_str, err := json.Marshal(pvc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddPVC] Failed to marshal data, " + err.Error(),
		})
		return
	}

	err = s.EtcdWrap.Put(e_key, pvc_str)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddPVC] Failed to write into etcd, " + err.Error(),
		})
		return
	}

	//返回200
	c.JSON(http.StatusOK, gin.H{
		"data": "Create pvc success.",
	})
}

func (s *ApiServer) DeletePVC(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/DeletePVC] Empty PVC name.",
		})
		return
	}

	e_key := apiserver.ETCD_pvc_prefix + name
	err := s.EtcdWrap.Del(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/DeletePVC] Failed to delete from etcd, " + err.Error(),
		})
		return
	}

	//返回200
	c.JSON(http.StatusOK, gin.H{
		"data": "Delete pvc success.",
	})
}
