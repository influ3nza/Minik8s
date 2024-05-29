package function

import (
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/network"
	"minik8s/tools"
	"os"
	"strconv"
	"sync"
	"time"
)

type Record struct {
	Name      string `json:"name" yaml:"name"`
	CallCount int32  `json:"callCount" yaml:"callcount"`
	Replicas  int32  `json:"replicas"  yaml:"replicas"`
}

var (
	RecordMap   = make(map[string]*Record)
	RecordMutex sync.RWMutex
)

//-----------------在trigger后调用----------------------//

func (fc *FunctionController) UpdateFunction(f *api_obj.Function) {
	// 获取写锁以防止并发写入
	RecordMutex.Lock()
	defer RecordMutex.Unlock()

	// 检查是否已经存在该记录
	if _, exists := RecordMap[f.Metadata.Name]; !exists {
		// 如果不存在，创建新的记录并添加到 RecordMap
		newRecord := &Record{
			Name:      f.Metadata.Name,
			CallCount: 100,
			Replicas:  3,
		}
		RecordMap[f.Metadata.Name] = newRecord
	} else {
		RecordMap[f.Metadata.Name].CallCount += int32(50 / (RecordMap[f.Metadata.Name].Replicas + 1))
	}

	if RecordMap[f.Metadata.Name].CallCount > 150 {
		tools.Scale_RS_Lock = true
	}
	_ = appendIntToFile("/mydata/record/f.txt", int(RecordMap[f.Metadata.Name].CallCount))
}

//-------------------------------以下是协程的内容----------------------------------//

func (fc *FunctionController) watch() {
	// 获取写锁以防止并发写入
	RecordMutex.Lock()
	defer RecordMutex.Unlock()

	for _, record := range RecordMap {
		if record.CallCount > 150*(record.Replicas) {
			replica, err := fc.scaleup(record)
			record.CallCount = 100
			tools.Scale_RS_Lock = false
			if err != nil {
				fmt.Println("Send Get RequestErr in watch ", err.Error())
				return
			}
			record.Replicas = int32(replica)
		} else if record.CallCount == 0 {
			replica, err := fc.scaledown(record)
			record.CallCount = 100
			if err != nil {
				fmt.Println("Send Get RequestErr in watch ", err.Error())
				return
			}
			record.Replicas = int32(replica)
		} else {
			record.CallCount -= 1
		}

		_ = appendIntToFile("/mydata/record/f.txt", int(record.CallCount))
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
	replica, err := strconv.Atoi(replicaStr)
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
	replica, err := strconv.Atoi(replicaStr)
	return replica, nil
}

func (fc *FunctionController) RunWatch() {
	go func() {
		for {
			go fc.watch()           // 执行 watch() 函数
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
