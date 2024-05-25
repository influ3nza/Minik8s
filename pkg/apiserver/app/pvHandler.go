package app

import (
	"encoding/json"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/config/kubelet"
	"minik8s/pkg/network"
	"minik8s/tools"
	"net/http"
	"os/exec"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *ApiServer) AddPV(c *gin.Context) {
	// TODO
	pv := &api_obj.PV{}
	err := c.ShouldBind(pv)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/AddPV] Failed to parse pv, " + err.Error(),
		})
		return
	}

	name := pv.Metadata.Name
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/AddPV] Empty pv name.",
		})
		return
	}

	e_key := apiserver.ETCD_pv_prefix + name
	res, err := s.EtcdWrap.Get(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddPV] Failed to get from etcd, " + err.Error(),
		})
		return
	}

	if len(res) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/AddPV] PV name already exists.",
		})
		return
	}

	pv_str, err := json.Marshal(pv)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddPV] Failed to marshal data, " + err.Error(),
		})
		return
	}

	err = s.EtcdWrap.Put(e_key, pv_str)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddPV] Failed to write into etcd, " + err.Error(),
		})
		return
	}

	//本地创建文件夹
	_, _ = exec.Command("mkdir", tools.PV_mount_master_path+pv.Spec.Nfs.Path).CombinedOutput()

	//修改/etc/exports文件
	args := []string{"\"" + tools.PV_mount_master_path + pv.Spec.Nfs.Path + tools.PV_mount_params + "\"", ">>", tools.PV_mount_config_file}
	_, err = exec.Command("echo", args...).CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/AddPV] Failed to update exports file, " + err.Error(),
		})
		return
	}

	_, _ = exec.Command("exportfs", "-r").CombinedOutput()

	//向所有node发送绑定消息
	for _, ip := range tools.NodesIpMap {
		uri := ip + strconv.Itoa(int(kubelet.Port)) + kubelet.MountNfs
		_, err := network.PostRequest(uri, pv_str)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "[ERR/handler/AddPV] Failed to send POST request, " + err.Error(),
			})
			return
		}
	}

	//返回200
	c.JSON(http.StatusOK, gin.H{
		"data": "Create pv success.",
	})
}

func (s *ApiServer) DeletePV(c *gin.Context) {
	// TODO
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/DeletePV] Empty PV name.",
		})
		return
	}

	e_key := apiserver.ETCD_pv_prefix + name
	err := s.EtcdWrap.Del(e_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/DeletePV] Failed to delete from etcd, " + err.Error(),
		})
		return
	}

	//TODO:向node发送删除请求
	//返回200
	c.JSON(http.StatusOK, gin.H{
		"data": "Delete pv success.",
	})
}
