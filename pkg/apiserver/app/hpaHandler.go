package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/config/apiserver"
)

func (s *ApiServer) AddHPA(c *gin.Context) {
	fmt.Printf("[apiserver/AddHPA] Try to add an hpa.\n")

	new_hpa := &api_obj.HPA{}
	err := c.ShouldBind(new_hpa)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/AddHPA] Failed to parse hpa, " + err.Error(),
		})
		return
	}

	h_key := apiserver.ETCD_hpa_prefix + new_hpa.MetaData.NameSpace + "/" + new_hpa.MetaData.Name
	h_str, err := json.Marshal(new_hpa)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/hpahandler/AddHPA] Failed to marshal hpa, " + err.Error(),
		})
		return
	}

	err = s.EtcdWrap.Put(h_key, h_str)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/hpahandler/AddHPA] Failed to write into etcd, " + err.Error(),
		})
		return
	}

	//返回200
	c.JSON(http.StatusOK, gin.H{
		"data": "[hpahandler/AddHPA] Add hpa success",
	})

}

func (s *ApiServer) DeleteHPA(c *gin.Context) {
	fmt.Printf("[apiserver/DeleteHPA] Try to delete HPA.\n")

	namespace := c.Param("namespace")
	name := c.Param("name")
	if name == "" || namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/HPAhandler/DeleteHPA] Service name and namespace shall not be null.",
		})
		return
	}

	fmt.Printf("[HPAhandler/DeleteHPA] namespace: %s, name: %s\n", name, namespace)
	err := s.EtcdWrap.Del(apiserver.ETCD_hpa_prefix + namespace + "/" + name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/HPAhandler/DeleteHPA] Failed to delete from etcd, " + err.Error(),
		})
		return
	}

	//返回200
	c.JSON(http.StatusOK, gin.H{
		"data": "[HPAhandler/DeleteHPA] Delete hpa success",
	})
}

func (s *ApiServer) GetHPAs(c *gin.Context) {
	fmt.Printf("[apiserver/GetHPA] Try to get HPAs.\n")

	key := apiserver.ETCD_hpa_prefix
	hpas, err := s.EtcdWrap.GetByPrefix(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/hpahandler/GetHPAs] Failed to get from etcd, " + err.Error(),
		})
		return
	}

	var allhpa []string
	for id, hpaItr := range hpas {
		allhpa = append(allhpa, hpaItr.Value)

		//返回值以逗号隔开
		if id < len(hpas)-1 {
			allhpa = append(allhpa, ",")
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": allhpa,
	})
}

func (s *ApiServer) GetHPA(c *gin.Context) {
	fmt.Printf("[apiserver/GetHPA] Try to get an hpa.\n")

	hpaname := c.Param("name")
	namespace := c.Param("namespace")

	if hpaname == "" || namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[ERR/handler/GetEndpoint] Endpoint name shall not be null.",
		})
		return
	}

	hpa_key := apiserver.ETCD_hpa_prefix + namespace + "/" + hpaname
	res, err := s.EtcdWrap.Get(hpa_key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/GetHPA] Failed to get from etcd, " + err.Error(),
		})
		return
	}

	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[ERR/handler/GetHPA] Found zero or more than one endpoint.\n",
		})
		return
	}

	var arr []string
	arr = append(arr, res[0].Value)

	//返回200
	c.JSON(http.StatusOK, gin.H{
		"data": arr,
	})
}
