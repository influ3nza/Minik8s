package object

import "sync"

type Record struct {
	Replicas  int32
	PodsIpCnt map[string]int32
	Count     int32
	Lock      *sync.Mutex
}

var AllRecords map[string]Record
