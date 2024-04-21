package etcd

import (
	"fmt"
	"testing"
	"time"
)

var etcdClientDummy *EtcdWrap = nil

func TestMain(m *testing.M) {
	wrap, err := CreateEtcdInstance([]string{"localhost:2379"}, 5*time.Second)
	if err != nil {
		_ = fmt.Errorf("[ERR/etcd_test/main] Failed to create etcd instance.\n")
		return
	}
	etcdClientDummy = wrap
	etcdClientDummy.DelAll()
	m.Run()
}

func TestPut(t *testing.T) {
	err := etcdClientDummy.Put("/test/word1", []byte("this is word 1"))
	if err != nil {
		t.Errorf("[ERR/etcd_test/Put] Failed to put a word\n")
		return
	}

	err = etcdClientDummy.Put("/test/word2", []byte("this is word 2"))
	if err != nil {
		t.Errorf("[ERR/etcd_test/Put] Failed to put a word\n")
		return
	}

	err = etcdClientDummy.Put("/test/word3", []byte("this is word 3"))
	if err != nil {
		t.Errorf("[ERR/etcd_test/Put] Failed to put a word\n")
		return
	}

	fmt.Printf("[PASS/etcd_test/Put]\n")
}

func TestGet(t *testing.T) {
	kv, err := etcdClientDummy.Get("/test/word1")
	if err != nil {
		t.Errorf("[ERR/etcd_test/Get] Failed to get a word\n")
		return
	}
	if kv.Value != "this is word 1" {
		t.Errorf("[ERR/etcd_test/Get] Failed to get the correct word\n")
		return
	}

	kv, err = etcdClientDummy.Get("/test")
	if err != nil {
		t.Errorf("[ERR/etcd_test/Get] Failed to get words\n")
		return
	}
	if kv.Version != -1 {
		t.Errorf("[ERR/etcd_test/Get] Fetched word should be null\n")
		return
	}

	fmt.Printf("[PASS/etcd_test/Get]\n")
}

func TestDel(t *testing.T) {
	err := etcdClientDummy.Del("/test/word1")
	if err != nil {
		t.Errorf("[ERR/etcd_test/Del] Failed to delete a word\n")
		return
	}

	kv, err := etcdClientDummy.Get("/test/word1")
	if err != nil {
		t.Errorf("[ERR/etcd_test/Del] Failed to get a word\n")
		return
	}
	if kv.Version != 1 {
		t.Errorf("[ERR/etcd_test/Del] Fetched word should be null\n")
		return
	}

	fmt.Printf("[PASS/etcd_test/Del]\n")
}
