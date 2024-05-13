package util

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

var wg sync.WaitGroup

func generateTest() map[string]string {
	var a map[string]string
	prefix := "test"
	for i := 1; i < 17; i++ {
		test := prefix + strconv.Itoa(i)
		a[test] = test
	}
	return a
}

func TestLock(t *testing.T) {
	// keys := generateTest()
	lockin := func(str string) {
		RegisterPod(str, str)
		for {
			res := Lock(str, str)
			if res == false {
				break
			}
			res = UnLock(str, str)
			if res == false {
				fmt.Println("Lock But Unlock Error, Means Implement Wrong")
			}
		}
		wg.Done()
	}

	lockout := func(str string) {
		time.Sleep(3 * time.Second)
		for {
			if res := UnRegisterPod(str, str); res == true {
				break
			}
		}
		wg.Done()
	}
	wg.Add(8)
	for i := 1; i < 5; i++ {
		test_ := "test" + strconv.Itoa(i)
		go lockin(test_)
		go lockout(test_)
	}
	wg.Wait()
}
