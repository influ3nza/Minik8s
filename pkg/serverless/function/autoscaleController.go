package function

import (
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/network"
	"os"
	"strconv"
	"sync"
	"time"
)

type Record struct {
	Name      string `json:"name" yaml:"name"`
	FuncTion  *api_obj.Function
	CallCount int32 `json:"callCount" yaml:"callcount"`
	Replicas  int32 `json:"replicas"  yaml:"replicas"`
	IpMap     map[string]int
	Mutex     sync.RWMutex
}

var (
	RecordMap   = make(map[string]*Record)
	RecordMutex sync.RWMutex
)

//-----------------在trigger后调用----------------------//

func (fc *FunctionController) UpdateFunction(funcName string) {
	// 获取写锁以防止并发写入

	// RecordMutex.Lock()
	// // 检查是否已经存在该记录
	// if _, exists := RecordMap[f.Metadata.Name]; !exists {
	// 	// 如果不存在，创建新的记录并添加到 RecordMap
	// 	newRecord := &Record{
	// 		Name:      f.Metadata.Name,
	// 		CallCount: 100,
	// 		Replicas:  3,
	// 		IpMap:     map[string]int{},
	// 		Mutex:     sync.RWMutex{},
	// 	}
	// 	RecordMap[f.Metadata.Name] = newRecord
	// }
	// RecordMutex.Unlock()
	RecordMutex.RLock()
	defer RecordMutex.RUnlock()
	if _, exists := RecordMap[funcName]; !exists {
		fmt.Printf("RecordMap[%s] not exists", funcName)
		return
	}
	RecordMap[funcName].Mutex.Lock()
	defer RecordMap[funcName].Mutex.Unlock()
	if RecordMap[funcName].Replicas == 0 {
		RecordMap[funcName].Replicas = 2
	}
	RecordMap[funcName].CallCount += 30 / (RecordMap[funcName].Replicas + 2)
	if RecordMap[funcName].CallCount > 200 {
		replica, err := fc.scaleup(RecordMap[funcName])
		if err != nil {
			fmt.Println("Send Get RequestErr in UpdateFunction ", err.Error())
			return
		}
		RecordMap[funcName].Replicas = int32(replica)
		RecordMap[funcName].CallCount = 90
	}

}

//-------------------------------以下是协程的内容----------------------------------//

func (fc *FunctionController) watch() {
	// 获取写锁以防止并发写入
	RecordMutex.RLock()
	defer RecordMutex.RUnlock()

	for _, record := range RecordMap {
		record.Mutex.Lock()
		fmt.Println("watch", record.Name, record.CallCount, record.Replicas)
		if record.CallCount > 0 {
			record.CallCount -= 1
		}
		if record.CallCount <= 0 {
			res, err := GetFunctionPodIps(record.FuncTion, false)
			if err != nil {
				fmt.Println("GetFunctionPodIps err", err.Error())
				record.Mutex.Unlock()
				continue
			}

			record.Replicas = int32(len(res))
			if record.Replicas != 0 {
				replica, err := fc.scaledown(record)
				record.CallCount = 90
				if err != nil {
					fmt.Println("Send Get RequestErr in watch ", err.Error())
					record.Mutex.Unlock()
					continue
				}
				record.Replicas = int32(replica)
			}
		}
		record.Mutex.Unlock()
	}

}

func (fc *FunctionController) scaleup(record *Record) (int, error) {
	name := record.Name
	uri := apiserver.API_server_prefix + apiserver.API_scale_replicaset_prefix + name + "/add"
	replicaStr, err := network.GetRequest(uri)
	if err != nil {
		fmt.Println("Send Get RequestErr in scaleup", err.Error())
		return 0, err
	}
	replica, _ := strconv.Atoi(replicaStr)
	return replica, nil
}

func (fc *FunctionController) scaledown(record *Record) (int, error) {
	name := record.Name
	uri := apiserver.API_server_prefix + apiserver.API_scale_replicaset_prefix + name + "/minus"
	replicaStr, err := network.GetRequest(uri)
	if err != nil {
		fmt.Println("Send Get RequestErr in scaleup", err.Error())
		return 0, err
	}
	replica, _ := strconv.Atoi(replicaStr)
	return replica, nil
}

func (fc *FunctionController) RunWatch() {
	go func() {
		for {
			fc.watch()              // 执行 watch() 函数
			time.Sleep(time.Second) // 等待一秒钟
		}
	}()
}

func appendIntToFile(filename string, num int) error {
	// 打开文件，使用追加模式和写模式，如果文件不存在则创建
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// 将 int 转换为字符串
	str := strconv.Itoa(num)

	// 将字符串写入文件
	_, err = file.WriteString(str + "\n")
	if err != nil {
		return err
	}

	return nil
}
