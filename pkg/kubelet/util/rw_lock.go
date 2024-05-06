package util

import (
	"fmt"
	"sync"
)

var manageLocks = sync.Mutex{}
var podLocks = sync.Map{}

func generateKey(podName string, namespace string) string {
	return fmt.Sprintf("%s:%s", namespace, podName)
}

func RegisterPod(podName string, namespace string) {
	manageLocks.Lock()
	mtx := &sync.Mutex{}
	key := generateKey(podName, namespace)
	podLocks.Store(key, mtx)
	manageLocks.Unlock()
}

func UnRegisterPod(podName string, namespace string) int {
	key := generateKey(podName, namespace)
	manageLocks.Lock()
	defer manageLocks.Unlock()
	mutex, ok := podLocks.Load(key)
	if ok {
		res := false
		for i := 0; i < 100; i++ {
			res = mutex.(*sync.Mutex).TryLock()
			if res == true {
				break
			}
		}
		if !res {
			return 1
		}
	} else {
		return 2 // not found
	}
	podLocks.Delete(key)
	mutex.(*sync.Mutex).Unlock()
	return 0
}

func Lock(podName string, namespace string) bool {
	key := generateKey(podName, namespace)
	manageLocks.Lock()
	mutex, ok := podLocks.Load(key)
	defer manageLocks.Unlock()
	if ok {
		mutex.(*sync.Mutex).Lock()
		return true
	} else {
		return false
	}
}

func UnLock(podName string, namespace string) bool {
	key := generateKey(podName, namespace)
	manageLocks.Lock()
	mutex, ok := podLocks.Load(key)
	defer manageLocks.Unlock()
	if ok {
		mutex.(*sync.Mutex).Unlock()
		return true
	} else {
		return false
	}
}
