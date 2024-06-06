package controller

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/network"
)

type callback func()

type HPAController struct {
}

func (hc *HPAController) PrintHandlerWarning() {
	fmt.Printf("[WARN/HPAController] Error in message handler, the system may not be working properly!\n")
}

var (
	hpatimedelay    = 5 * time.Second
	hpatimeinterval = []time.Duration{15 * time.Second}
)

func CreateHPAControllerInstance() (*HPAController, error) {
	return &HPAController{}, nil
}

func (hc *HPAController) Run() {
	hc.execute(hpatimedelay, hpatimeinterval, hc.watch)
}

func (hc *HPAController) execute(delay time.Duration, interval []time.Duration, callback callback) {
	if len(interval) == 0 {
		return
	}
	<-time.After(delay)
	for {
		for _, inter := range interval {
			callback()
			<-time.After(inter)
		}
	}
}

func (hc *HPAController) GetAllHPAs() ([]api_obj.HPA, error) {
	uri := apiserver.API_server_prefix + apiserver.API_get_hpas
	dataStr, err := network.GetRequest(uri)
	if err != nil {
		fmt.Printf("[ERR/HPAController/GetAllHPAs] GET request failed, %v.\n", err)
		hc.PrintHandlerWarning()
		return nil, err
	}

	var hpas []api_obj.HPA
	if dataStr == "" {
		fmt.Printf("[ERR/HPAController/GetAllHPAs] Not any hpa available.\n")
		hc.PrintHandlerWarning()
		return nil, err
	} else {
		err = json.Unmarshal([]byte(dataStr), &hpas)
		if err != nil {
			fmt.Printf("[ERR/HPAController/GetAllHPAs] Failed to unmarshal data, %s.\n", err)
			hc.PrintHandlerWarning()
			return nil, err
		}
		return hpas, nil
	}
}

func (hc *HPAController) watch() {
	//返回所有的pods
	uri := apiserver.API_server_prefix + apiserver.API_get_pods
	dataStr, err := network.GetRequest(uri)
	if err != nil {
		fmt.Printf("[ERR/HPAController/watch] GET request failed, %v.\n", err)
		hc.PrintHandlerWarning()
		return
	}
	var allPods []api_obj.Pod
	if dataStr == "" {
		fmt.Printf("[ERR/HPAController/watch] Not any pod available.\n")
		hc.PrintHandlerWarning()
	} else {
		err = json.Unmarshal([]byte(dataStr), &allPods)
		if err != nil {
			fmt.Printf("[ERR/HPAController/watch] Failed to unmarshal data, %s.\n", err)
			hc.PrintHandlerWarning()
			return
		}
	}

	//返回所有的hpa
	hpas, err := hc.GetAllHPAs()
	if err != nil {
		return
	}

	for _, hpa := range hpas {
		correspondPods := make([]api_obj.Pod, 0)
		for _, pod := range allPods {
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
		// 先直接扩容到expectreplica
		if hpa.Status.CurReplicas < expectedReplicas {
			err := hc.AddHpaPod(hpa, correspondPods, 1)
			if err != nil {
				fmt.Printf("[ERR/HPAController/watch] Failed to addhpapod, %s.\n", err)
				hc.PrintHandlerWarning()
				return
			}
		}

		// 根据策略缩容
		// 先直接扩容到expectreplica
		if hpa.Status.CurReplicas > expectedReplicas {
			err := hc.ReduceHpaPod(hpa, correspondPods, hpa.Status.CurReplicas-expectedReplicas)
			if err != nil {
				fmt.Printf("[ERR/HPAController/watch] Failed to reducehpapod, %s.\n", err)
				hc.PrintHandlerWarning()
				return
			}
		}
		// d.更新replica的个数
		hpa.Status.CurReplicas = expectedReplicas

		// 3. 判断hpa的pod是否在规定区间之内, 如果不在，需要扩容或者缩容
		// a.根据策略扩容
		// 如果是policy是Pods就扩容到最小，如果policy是Percent就扩容到(min+max)/2
		if hpa.Status.CurReplicas < hpa.Spec.MinReplicas {
			if *hpa.Spec.Policy == api_obj.PodsPolicy || hpa.Spec.Policy == nil {
				err := hc.AddHpaPod(hpa, correspondPods, hpa.Spec.MinReplicas-hpa.Status.CurReplicas)
				if err != nil {
					fmt.Printf("[ERR/HPAController/watch] Failed to addhpapod in pods, %s.\n", err)
					hc.PrintHandlerWarning()
					return
				}

			} else if *hpa.Spec.Policy == api_obj.PercentPolicy {
				err := hc.AddHpaPod(hpa, correspondPods, (hpa.Spec.MinReplicas+hpa.Spec.MaxReplicas)/2-hpa.Status.CurReplicas)
				if err != nil {
					fmt.Printf("[ERR/HPAController/watch] Failed to addhpapod in percent, %s.\n", err)
					hc.PrintHandlerWarning()
					return
				}
			}
		}

		// b.根据策略缩容
		// 如果是policy是Pods就缩容到最小，如果policy是Percent就扩容到(min+max)/2
		if hpa.Status.CurReplicas > hpa.Spec.MaxReplicas {
			if *hpa.Spec.Policy == api_obj.PodsPolicy || hpa.Spec.Policy == nil {
				err := hc.ReduceHpaPod(hpa, correspondPods, hpa.Status.CurReplicas-hpa.Spec.MaxReplicas)
				if err != nil {
					fmt.Printf("[ERR/HPAController/watch] Failed to reducehpapod in pod, %s.\n", err)
					hc.PrintHandlerWarning()
					return
				}
			} else if *hpa.Spec.Policy == api_obj.PercentPolicy {
				err := hc.ReduceHpaPod(hpa, correspondPods, hpa.Status.CurReplicas-(hpa.Spec.MinReplicas+hpa.Spec.MaxReplicas)/2)
				if err != nil {
					fmt.Printf("[ERR/HPAController/watch] Failed to reducehpapod in percent, %s.\n", err)
					hc.PrintHandlerWarning()
					return
				}
			}
		}

		// 4.更新hpa状态
		hpa.Status.CurCpu = averageCPUUsage
		hpa.Status.CurMemory = averageMemoryUsage
		err := hc.UpdateHpa(hpa)
		if err != nil {
			fmt.Printf("[ERR/HPAController/watch] Failed to updateHpa, %s.\n", err)
			hc.PrintHandlerWarning()
			return
		}
	}

}

func (hc *HPAController) AddHpaPod(hpa api_obj.HPA, pods []api_obj.Pod, num int) error {
	uri := apiserver.API_server_prefix + apiserver.API_add_pod
	podNew := api_obj.Pod{}
	podNew.MetaData = pods[0].MetaData
	podNew.ApiVersion = "v1"
	podNew.Kind = "Pod"
	podNew.Spec = pods[0].Spec
	podNew.Spec.NodeName = ""
	podNew.MetaData.Name = hpa.Spec.Workload.Name
	podNew.MetaData.NameSpace = hpa.Spec.Workload.NameSpace
	podNew.MetaData.Labels["hpa_name"] = hpa.MetaData.Name
	podNew.MetaData.Labels["hpa_namespace"] = hpa.MetaData.NameSpace
	podNew.MetaData.Labels["hpa_uuid"] = hpa.MetaData.UUID

	podName := podNew.MetaData.Name

	for i := 0; i < num; i++ {
		rand.Seed(time.Now().UnixNano())
		randomNumber := rand.Intn(1000)
		randomString := strconv.Itoa(randomNumber)
		podNew.MetaData.Name = podName + "-" + randomString + "hpaCreate" + strconv.Itoa(num)

		podJson, err := json.Marshal(podNew)
		if err != nil {
			fmt.Printf("[ERR/hpaController/AddHpaPod] Failed to marshal pod, %v.\n", err)
			return err
		}

		_, err = network.PostRequest(uri, podJson)
		if err != nil {
			fmt.Printf("[ERR/hpaController/AddHpaPod] Failed to post request, err:%v\n", err)
			return err
		}

	}
	fmt.Printf("[hpaController/AddHpaPod] Send add pod request success!\n")

	return nil
}

func (hc *HPAController) ReduceHpaPod(hpa api_obj.HPA, pods []api_obj.Pod, num int) error {
	for i := 0; i < num; i++ {
		namespace := pods[i].MetaData.NameSpace
		name := pods[i].MetaData.Name
		uri := apiserver.API_server_prefix + apiserver.API_delete_pod_prefix + namespace + "/" + name
		_, err := json.Marshal(pods[i])
		if err != nil {
			fmt.Printf("[ERR/hpaController/ReduceHpaPod] Failed to marshal pod, %v.\n", err)
			return err
		}

		_, err = network.DelRequest(uri)
		if err != nil {
			fmt.Printf("[ERR/hpaController/ReduceHpaPod] Failed to post request, err:%v\n", err)
			return err
		}
	}
	fmt.Printf("[hpaController/ReduceHpaPod] Send delete pod request success!\n")
	return nil
}

func (hc *HPAController) UpdateHpa(hpa api_obj.HPA) error {
	uri := apiserver.API_server_prefix + apiserver.API_update_hpa
	newHpa := api_obj.HPA{}

	newHpa = hpa
	hpaStr, err := json.Marshal(newHpa)
	if err != nil {
		fmt.Printf("[ERR/hpaController/UpdateHpaStatus] Failed to marshal hpa, %v.\n", err)
		return err
	}
	_, err = network.PostRequest(uri, hpaStr)
	if err != nil {
		fmt.Printf("[ERR/hpaController/UpdateHpaStatus] Failed to post request, err:%v\n", err)
		return err
	}
	return nil
}

func CheckPod(pod *api_obj.Pod, selectors map[string]string) bool {
	podLabel := pod.MetaData.Labels
	//may have bug
	for key, value := range selectors {
		if podLabel[key] == value {
			return true
		}
	}
	return false
}

// 通过for循环遍历每个pod，对每个pod通过name和namespace构造uri来调用getpodmetrics，返回该pod的
// containermetrics数组，对这个数组求和得到每一个pod的CPUPercent和MemoryPercent
// 取平均得到平均数
func (hc *HPAController) AverageCPUUsage(pods []api_obj.Pod) float64 {
	var sum float64 = 0.00
	length := len(pods)
	for _, pod := range pods {
		namespace := pod.MetaData.NameSpace
		name := pod.MetaData.Name
		uri := apiserver.API_server_prefix + apiserver.API_get_pod_metrics_prefix + namespace + "/" + name
		dataStr, err := network.GetRequest(uri)
		if err != nil {
			fmt.Printf("[ERR/EndpointController/OnAddService] GET request failed, %v.\n", err)
			hc.PrintHandlerWarning()
		}
		podMetrics := &api_obj.PodMetrics{}
		err = json.Unmarshal([]byte(dataStr), podMetrics)
		if err != nil {
			fmt.Printf("[ERR/HPAController/AverageCPUUsage] Failed to unmarshal pod, " + err.Error())
			hc.PrintHandlerWarning()
		}

		fmt.Printf("[PodMetrics] %v\n", podMetrics)

		containerMetricss := podMetrics.ContainerMetrics
		var metrics float64 = 0.00
		for _, containerMetrics := range containerMetricss {
			fmt.Printf("[Calculate CPU usage] %2f\n", containerMetrics.Usage.CPUPercent)
			metrics += float64(containerMetrics.Usage.CPUPercent)
		}
		sum += metrics
	}
	sum = sum / float64(length)
	return sum
}

func (hc *HPAController) AverageMemoryUsage(pods []api_obj.Pod) float64 {
	var sum float64 = 0.00
	length := len(pods)
	for _, pod := range pods {
		namespace := pod.MetaData.NameSpace
		name := pod.MetaData.Name
		uri := apiserver.API_server_prefix + apiserver.API_get_pod_metrics_prefix + namespace + "/" + name
		dataStr, err := network.GetRequest(uri)
		if err != nil {
			fmt.Printf("[ERR/EndpointController/OnAddService] GET request failed, %v.\n", err)
			hc.PrintHandlerWarning()
		}
		podMetrics := &api_obj.PodMetrics{}
		err = json.Unmarshal([]byte(dataStr), podMetrics)
		if err != nil {
			fmt.Printf("[ERR/HPAController/AverageCPUUsage] Failed to unmarshal pod, " + err.Error())
			hc.PrintHandlerWarning()
		}

		containerMetricss := podMetrics.ContainerMetrics
		var metrics float64 = 0.00
		for _, containerMetrics := range containerMetricss {
			metrics += float64(containerMetrics.Usage.MemoryPercent)
		}
		sum += metrics
	}
	sum = sum / float64(length)
	return sum
}

func (hc *HPAController) ExpectedReplicas(hpa api_obj.HPA, averageCPUUsage float64, averageMemoryUsage float64) int {
	cpuUsedPropotion := averageCPUUsage / hpa.Spec.Metrics.CPUPercent
	memoryUsedPropotion := averageMemoryUsage / hpa.Spec.Metrics.MemPercent
	expectedReplicas := int(math.Max(cpuUsedPropotion, memoryUsedPropotion) * float64(hpa.Status.CurReplicas))

	fmt.Printf("[INFO/HPAController/ExpectedReplicas] CPU&MEM threshold: %2f, %2f\n", hpa.Spec.Metrics.CPUPercent, hpa.Spec.Metrics.MemPercent)
	fmt.Printf("[INFO/HPAController/ExpectedReplicas] cpuUsedPropotion: %.2f\n", cpuUsedPropotion)
	fmt.Printf("[INFO/HPAController/ExpectedReplicas] memoryUsedPropotion: %.2f\n", memoryUsedPropotion)
	fmt.Printf("[INFO/HPAController/ExpectedReplicas] expectedReplicas: %d\n", expectedReplicas)

	return int(math.Max(float64(expectedReplicas), float64(hpa.Spec.MinReplicas)))
}
