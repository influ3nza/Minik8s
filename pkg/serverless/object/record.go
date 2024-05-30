package object

import (
	"minik8s/pkg/api_obj"
	"sync"
)

type Record struct {
	Function  *api_obj.Function
	Replicas  int32
	PodsIpCnt map[string]int32
	Count     int32
	Lock      *sync.Mutex
}

var AllRecords map[string]Record
