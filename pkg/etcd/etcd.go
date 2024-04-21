package etcd

import (
	"context"
	"fmt"
	"time"

	etcd "go.etcd.io/etcd/client/v3"
)

type EtcdWrap struct {
	client *etcd.Client
}

type EtcdKV struct {
	Version int64
	Key     string
	Value   string
}

func CreateEtcdInstance(endpoints []string, dialTimeout time.Duration) (*EtcdWrap, error) {
	cli, err := etcd.New(etcd.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})

	if err != nil {
		fmt.Printf("connect to etcd failed, err:%v\n", err)
		return nil, err
	}
	fmt.Println("connect to etcd success")

	//这里不用defer，client在函数返回之后还需要继续使用。

	return &EtcdWrap{client: cli}, nil
}

func (w *EtcdWrap) Put(key string, val []byte) error {
	_, err := w.client.Put(context.TODO(), key, string(val))
	return err
}

func (w *EtcdWrap) Get(key string) (EtcdKV, error) {
	resp, err := w.client.Get(context.TODO(), key)

	if err != nil || len(resp.Kvs) == 0 {
		return EtcdKV{Version: -1}, err
	}

	ev := resp.Kvs[0]
	return EtcdKV{
		Version: ev.Version,
		Key:     string(ev.Key),
		Value:   string(ev.Value)}, nil
}

func (w *EtcdWrap) Del(key string) error {
	_, err := w.client.Delete(context.TODO(), key)
	return err
}

func (w *EtcdWrap) DelAll() error {
	_, err := w.client.Delete(context.TODO(), "", etcd.WithPrefix())
	return err
}

func (w *EtcdWrap) GetByPrefix(key string) ([]EtcdKV, error) {
	resp, err := w.client.Get(context.TODO(), key, etcd.WithPrefix())
	if err != nil {
		return []EtcdKV{}, err
	}

	var pack []EtcdKV
	for id, kv := range resp.Kvs {
		pack = append(pack, EtcdKV{
			Version: resp.Kvs[id].Version,
			Key:     string(kv.Key),
			Value:   string(kv.Value),
		})
	}

	return pack, nil
}
