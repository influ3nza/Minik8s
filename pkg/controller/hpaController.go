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

func (hc *HPAController) GetAllHPAs() ([]api_obj.HPA, error) {
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

func (hc *HPAController) watch() {
	//返回所有的pods
	uri := config.API_server_prefix + config.API_get_pods
	dataStr, err := network.GetRequest(uri)
	if err != nil {
		fmt.Printf("[ERR/HPAController/watch] GET request failed, %v.\n", err)
		ec.PrintHandlerWarning()
		return
	}
	var allPods []api_obj.Pod
	if dataStr == "" {
		fmt.Printf("[ERR/HPAController/watch] Not any pod available.\n")
		ec.PrintHandlerWarning()
	} else {
		err = json.Unmarshal([]byte(dataStr), &allPods)
		if err != nil {
			fmt.Printf("[ERR/HPAController/watch] Failed to unmarshal data, %s.\n", err)
			ec.PrintHandlerWarning()
			return
		}
	}

	//返回所有的hpa
	hpas, err := rc.GetAllHPAs()
	if err != nil {
		return
	}

	for _, hpa := range hpas {
		correspondPods := make([]api_obj.Pod, 0)
		for _, pod := range pods {
			if CheckPod(&pod, hpa.Spec.Selector) {
				correspondPods = append(correspondPods, pod)
			}
		}
		hpa.Status.CurReplicas = len(correspondPods)
		// 2.根据资源来扩容或缩容
			// a. 计算hpa匹配的pod的平均cpu和memory使用率
		averageCPUUsage := hc.AverageCPUUsage(correspondPods)
		averageMemoryUsage := hc.AverageMemoryUsage(correspondPods)

			// b. 根据hpa的spec和计算出来的平均使用率，得到期望的replica个数
		expectedReplicas := hc.ExpectedReplicas(hpa, averageCPUUsage, averageMemoryUsage)

			// c. 判断hpa的pod是否在规定区间之内, 如果不在，需要扩容或者缩容
				// 根据策略扩容
				// 如果是policy是Pods就扩容到最小，如果policy是Percent就扩容到(min+max)/2
		if hpa.Status.CurReplicas < expectedReplicas {
			err := hc.AddHpaPod(hpa, correspondPods, expectedReplicas-hpa.Status.CurReplicas)
			if err != nil {
				k8log.ErrorLog("hpaController", "HandleHPAUpdate "+err.Error())
			}
			return
		}

				// 根据策略缩容
				// 如果是policy是Pods就缩容到最小，如果policy是Percent就扩容到(min+max)/2
		if hpa.Status.CurReplicas > expectedReplicas {
			err := hc.ReduceHpaPod(hpa, correspondPods)
			if err != nil {
				k8log.ErrorLog("hpaController", "HandleHPAUpdate "+err.Error())
			}
			return
		}
			//更新replica的个数
		hpa.Status.CurReplicas = expectedReplicas


		// 3. 判断hpa的pod是否在规定区间之内, 如果不在，需要扩容或者缩容
			// a.根据策略扩容
			// 如果是policy是Pods就扩容到最小，如果policy是Percent就扩容到(min+max)/2
		if hpa.Status.CurReplicas < hpa.Spec.MinReplicas {
			err := hc.AddHpaPod(hpa, correspondPods, hpa.Spec.MinReplicas-hpa.Status.CurReplicas)
			if err != nil {
				k8log.ErrorLog("hpaController", "HandleHPAUpdate "+err.Error())
			}
			return
		}

			// b.根据策略缩容
			// 如果是policy是Pods就缩容到最小，如果policy是Percent就扩容到(min+max)/2
		if hpa.Status.CurrentReplicas > hpa.Spec.MaxReplicas {
			err := hc.ReduceHpaPod(hpa, correspondPods, hpa.Status.CurrentReplicas-hpa.Spec.MaxReplicas)
			if err != nil {
				k8log.ErrorLog("hpaController", "HandleHPAUpdate "+err.Error())
			}
			return
		}

		// 4.更新hpa状态
		hpa.Status.CurCpu = averageCPUUsage
		hpa.Status.CurMemory = averageMemoryUsage
		err := hc.UpdateHpaStatus(hpa)
		if err != nil {
			k8log.ErrorLog("hpaController", "HandleHPAUpdate "+err.Error())
		}
	}

}

func (hc *HPAController) AddHpaPod(hpa *api_obj.HPA, pods []api_obj.Pod, num int) error {

}

func (hc *HPAController) ReduceHpaPod(hpa *api_obj.HPA, pods []api_obj.Pod, num int) error {
	
}

func (hc *HPAController) UpdateHpaStatus(hpa *api_obj.HPA) error {
	
}

//通过for循环遍历每个pod，对每个pod通过name和namespace构造uri来调用getpodmetrics，返回该pod的
//containermetrics数组，对这个数组求和得到每一个pod的CPUPercent和MemoryPercent
//取平均得到平均数
func (hc *HPAController) AverageCPUUsage(pods []api_obj.Pod) float64 {

}

func (hc *HPAController) AverageMemoryUsage(pods []api_obj.Pod) float64 {

}

func (hc *HPAController) ExpectedReplicas(hpa *api_obj.HPA, averageCPUUsage float64, averageMemoryUsage float64) int {

}


