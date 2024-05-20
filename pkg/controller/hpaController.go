package controller

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/controller/utils"
	"minik8s/pkg/message"
	"minik8s/pkg/network"
)

type HPAController struct {
}

func (rc *HPAController) GetAllHPAs() ([]api_obj.HPA, error) {
	uri := config.API_server_prefix + config.API_get_hpas
	dataStr, err := network.GetRequest(uri)
	if err != nil {
		fmt.Printf("[ERR/HPAController/GetAllHPAs] GET request failed, %v.\n", err)
		ec.PrintHandlerWarning()
		return
	}

	var hpas []api_obj.HPA
	if dataStr == "" {
		fmt.Printf("[ERR/HPAController/GetAllHPAs] Not any hpa available.\n")
		ec.PrintHandlerWarning()
	} else {
		err = json.Unmarshal([]byte(dataStr), &hpas)
		if err != nil {
			fmt.Printf("[ERR/HPAController/GetAllHPAs] Failed to unmarshal data, %s.\n", err)
			ec.PrintHandlerWarning()
			return nil,err
		}
		return hpas, nil
	}
}