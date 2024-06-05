package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/apiserver/controller/utils"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/network"
	"net/http"
	"sort"
	"sync"
	"time"
)

type FunctionController struct {
}

func (fc *FunctionController) GenerateFunction(f *api_obj.Function) error {
	err := CreateImage(f)
	if err != nil {
		return fmt.Errorf("create Img Failed At GF, %s", err.Error())
	}

	err = fc.GenerateReplicaset(f)
	if err != nil {
		return fmt.Errorf("gen replica Failed At GF, %s", err.Error())
	}

	RecordMutex.Lock()
	RecordMap[f.Metadata.Name] = &Record{
		Name:      f.Metadata.Name,
		FuncTion:  f,
		CallCount: 100,
		Replicas:  0,
		IpMap:     map[string]int{},
		Mutex:     sync.RWMutex{},
	}
	RecordMutex.Unlock()

	//是否通过监视文件触发（文件夹）
	if f.NeedWatch {
		Watcher.AddWatchDir("/mydata/" + f.Metadata.UUID + "/" + f.Metadata.Name + "/")
		fmt.Printf("[GenerateFunction] Add to file watch success, %s\n", f.Metadata.Name)
	}

	return nil
}

func CreateNewFunctionControllerInstance() *FunctionController {
	return &FunctionController{}
}

func (fc *FunctionController) GetFunctionPodIps(f *api_obj.Function) ([]string, error) {
	var array []string
	url := apiserver.API_server_prefix + apiserver.API_find_function_ip_prefix + f.Metadata.Name
	err := network.GetRequestAndParse(url, &array)
	if err != nil {
		fmt.Printf("get Func Pod Ips Failed, %s", err.Error())
		return []string{}, err
	}

	return array, nil
}

func (fc *FunctionController) GenerateReplicaset(f *api_obj.Function) error {
	imgName := serverDns + ":5000/" + f.Metadata.Name + ":latest"
	replica := &api_obj.ReplicaSet{
		ApiVersion: "v1",
		Kind:       "replicaset",
		MetaData: obj_inner.ObjectMeta{
			Name:      utils.RS_name_prefix + f.Metadata.Name,
			NameSpace: f.Metadata.NameSpace,
			Labels: map[string]string{
				"func": f.Metadata.Name,
			},
		},
		Spec: api_obj.ReplicaSetSpec{
			Replicas: 0,
			Selector: map[string]string{
				"func": f.Metadata.Name,
			},
			Template: api_obj.PodTemplate{
				Metadata: obj_inner.ObjectMeta{
					Name:      f.Metadata.Name, //todo need to be different
					NameSpace: f.Metadata.NameSpace,
					Labels: map[string]string{
						"func": f.Metadata.Name,
					},
				},
				Spec: api_obj.PodSpec{
					Containers: []api_obj.Container{
						{
							Name: f.Metadata.Name,
							Image: obj_inner.Image{
								Img:           imgName,
								ImgPullPolicy: "Always",
							},
						}, // one and only one container
					},
					Volumes:       nil,
					NodeName:      "",
					NodeSelector:  nil,
					RestartPolicy: "",
				},
			},
		},
		Status: api_obj.ReplicaSetStatus{
			Replicas:      0,
			ReadyReplicas: 0,
			Status:        "",
			Message:       "",
		},
	}

	replicaJson, err := json.Marshal(replica)
	if err != nil {
		return fmt.Errorf("marshal Replica Failed, %s", err.Error())
	}

	url := apiserver.API_server_prefix + apiserver.API_add_replicaset
	_, err = network.PostRequest(url, replicaJson)
	if err != nil {
		return fmt.Errorf("network Failed, %s", err.Error())
	}

	return nil
}

func (fc *FunctionController) MakePods(f *api_obj.Function) ([]string, error) {
	for {
		res, err := fc.GetFunctionPodIps(f)
		if err != nil {
			return []string{}, fmt.Errorf("get Pod Ips Failed, %s", err.Error())
		}

		if len(res) < 1 {
			time.Sleep(2 * time.Second)
		} else {
			return res, nil
		}
	}
}

func (fc *FunctionController) Schedule(funcName string, ips []string) string {
	if len(ips) > 0 {
		fmt.Println("ips are: ", ips)
		fmt.Println("Before Fr IpMap: ", RecordMap[funcName].IpMap)
		RecordMutex.RLock()
		fr := RecordMap[funcName]
		RecordMutex.RUnlock()
		if fr == nil {
			fmt.Printf("[Schedule] Function not found.\n")
			return ""
		}

		newIpMap := map[string]int{}
		for _, ip := range ips {
			if _, ok := fr.IpMap[ip]; !ok {
				newIpMap[ip] = 0
			} else {
				newIpMap[ip] = fr.IpMap[ip]
			}
		}
		fr.IpMap = newIpMap

		sort.Slice(ips, func(i, j int) bool {
			return fr.IpMap[ips[i]] < fr.IpMap[ips[j]]
		})

		ret := ips[0]
		fr.IpMap[ret] += 1
		fmt.Println("Now Fr IpMap: ", RecordMap[funcName].IpMap)
		return ret
	}

	return ""
}

func (fc *FunctionController) TriggerFunction(f *api_obj.Function) (string, error) {
	ips, err := fc.MakePods(f)
	if err != nil {
		return "", fmt.Errorf("get Ip Failed At Trigger, %s", err.Error())
	}

	ip := fc.Schedule(f.Metadata.Name, ips)
	if ip == "" {
		return "", fmt.Errorf("schedule Failed")
	}

	param := []byte(f.Coeff)
	// fmt.Println(param, f.Coeff)

	body := bytes.NewReader(param)
	uri := "http://" + ip + ":8080"
	failCnt := 0
	var resp *http.Response = nil
	for {
		resp, err = http.Post(uri, "application/json", body)
		if err != nil {
			failCnt += 1
			if failCnt >= 3 {
				return "", fmt.Errorf("post Req To Func Failed, %s", err.Error())
			}
			time.Sleep(2 * time.Second)
			continue
		}

		defer resp.Body.Close()
		break
	}

	var result map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&result)
	if err != nil {
		fmt.Printf("[FunctionExec] Error in returning http, %v", err)
	}
	fmt.Println(result)
	res, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal Result Failed %s", err.Error())
	}

	fmt.Println(string(res))

	return string(res), nil
}

func (fc *FunctionController) TriggerFunctionLocal(f *api_obj.Function, cof string) (string, error) {
	ips, err := fc.MakePods(f)
	if err != nil {
		return "", fmt.Errorf("get Ip Failed At TriggerLocal, %s", err.Error())
	}

	ip := fc.Schedule(f.Metadata.Name, ips)
	if ip == "" {
		return "", fmt.Errorf("schedule Failed")
	}

	param := []byte(cof)
	body := bytes.NewReader(param)
	uri := "http://" + ip + ":8080"
	failCnt := 0
	var resp *http.Response = nil
	for {
		resp, err = http.Post(uri, "application/json", body)
		if err != nil {
			failCnt += 1
			if failCnt >= 3 {
				return "", fmt.Errorf("post Req To Func Failed, %s", err.Error())
			}
			time.Sleep(2 * time.Second)
			continue
		}

		defer resp.Body.Close()
		break
	}

	var result map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&result)
	if err != nil {
		fmt.Printf("[FunctionExec] Error in returning http, %v", err)
	}
	fmt.Println(result)
	res, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal Result Failed %s", err.Error())
	}

	fmt.Println(string(res))

	return string(res), nil
}

func (fc *FunctionController) UpdateFunctionBody(f *api_obj.Function) error {
	err := fc.DeleteFunction(f)
	if err != nil {
		return fmt.Errorf("delete Replica Failed In Update Failed, %s", err.Error())
	}

	err = fc.GenerateFunction(f)
	if err != nil {
		return fmt.Errorf("generate Replica Failed In Update Failed, %s", err.Error())
	}
	return nil
}

func (fc *FunctionController) DeleteFunction(f *api_obj.Function) error {
	replicName := utils.RS_name_prefix + f.Metadata.Name
	url := apiserver.API_server_prefix +
		apiserver.API_delete_replicaset_prefix + apiserver.API_default_namespace + "/" + replicName
	_, err := network.DelRequest(url)
	if err != nil {
		return fmt.Errorf("send Delete Rep Failed, %s", err.Error())
	}

	RecordMutex.Lock()
	delete(RecordMap, f.Metadata.Name)
	RecordMutex.Unlock()
	//TODO:需要取消注释。
	err = DeleteImage(f.Metadata.Name)
	if err != nil {
		fmt.Printf("no Such Img or Delete Img error, %s", err.Error())
	}
	return nil
}
