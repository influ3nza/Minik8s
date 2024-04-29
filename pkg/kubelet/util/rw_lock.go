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

func UnRegisterPod(podName string, namespace string) {
	manageLocks.Lock()
	defer manageLocks.Unlock()
	key := generateKey(podName, namespace)
	mutex, ok := podLocks.Load(key)
	if ok {
		mutex.(*sync.Mutex).Lock()
	} else {
		return
	}
	podLocks.Delete(key)
	mutex.(*sync.Mutex).Unlock()
	return
}

func Lock(podName string, namespace string) bool {
	key := generateKey(podName, namespace)
	if mutex, ok := podLocks.Load(key); ok {
		mutex.(*sync.Mutex).Lock()
		return true
	} else {
		return false
	}
}

func UnLock(podName string, namespace string) bool {
	key := generateKey(podName, namespace)
	if mutex, ok := podLocks.Load(key); ok {
		mutex.(*sync.Mutex).Unlock()
		return true
	} else {
		return false
	}
}
