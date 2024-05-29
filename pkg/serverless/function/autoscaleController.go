package function

import (
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/network"
	"strconv"
	"sync"
	"time"
)

type Record struct {
	Name      string `json:"name" yaml:"name"`
	CallCount int32  `json:"callCount" yaml:"callcount"`
	Replicas  int32  `json:"replicas"  yaml:"replicas"`
	IpMap     map[string]int
	Mutex     sync.RWMutex
}

var (
	RecordMap   = make(map[string]*Record)
	RecordMutex sync.RWMutex
)

//-----------------在trigger后调用----------------------//

func (fc *FunctionController) UpdateFunction(f *api_obj.Function) {
	// 获取写锁以防止并发写入

	RecordMutex.Lock()
	// 检查是否已经存在该记录
	if _, exists := RecordMap[f.Metadata.Name]; !exists {
		// 如果不存在，创建新的记录并添加到 RecordMap
		newRecord := &Record{
			Name:      f.Metadata.Name,
			CallCount: 100,
			Replicas:  3,
			IpMap:     map[string]int{},
			Mutex:     sync.RWMutex{},
		}
		RecordMap[f.Metadata.Name] = newRecord
	}
	RecordMutex.Unlock()

	RecordMap[f.Metadata.Name].Mutex.Lock()
	RecordMap[f.Metadata.Name].CallCount += 120 / (RecordMap[f.Metadata.Name].Replicas + 1)
	RecordMap[f.Metadata.Name].Mutex.Unlock()
}

//-------------------------------以下是协程的内容----------------------------------//

func (fc *FunctionController) watch() {
	// 获取写锁以防止并发写入
	RecordMutex.RLock()
	defer RecordMutex.RUnlock()

	for _, record := range RecordMap {
		record.Mutex.RLock()
		if record.CallCount > 150 {
			replica, err := fc.scaleup(record)
			if err != nil {
				fmt.Println("Send Get RequestErr in watch ", err.Error())
				record.Mutex.RUnlock()
				continue
			}
			record.Replicas = int32(replica)
			record.Mutex.RUnlock()
		} else if record.CallCount == 0 {
			replica, err := fc.scaledown(record)
			if err != nil {
				fmt.Println("Send Get RequestErr in watch ", err.Error())
				record.Mutex.RUnlock()
				continue
			}
			record.Replicas = int32(replica)
			record.Mutex.RUnlock()
		} else {
			record.CallCount -= 2
			record.Mutex.RUnlock()
		}
	}
}

func (fc *FunctionController) scaleup(record *Record) (int, error) {
	name := record.Name
	uri := apiserver.API_scale_replicaset_prefix + name + "add"
	replicaStr, err := network.GetRequest(uri)
	if err != nil {
		fmt.Println("Send Get RequestErr in scaleup", err.Error())
		return 0, err
	}
	replica, err := strconv.Atoi(replicaStr)
	return replica, nil
}

func (fc *FunctionController) scaledown(record *Record) (int, error) {
	name := record.Name
	uri := apiserver.API_scale_replicaset_prefix + name + "minus"
	replicaStr, err := network.GetRequest(uri)
	if err != nil {
		fmt.Println("Send Get RequestErr in scaleup", err.Error())
		return 0, err
	}
	replica, err := strconv.Atoi(replicaStr)
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
