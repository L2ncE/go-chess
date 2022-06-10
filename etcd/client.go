package etcd

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"time"
)

type EClient struct {
	cli *clientv3.Client
}

type ConfigEtcdAddr struct {
	EtcdAddr string
}

type ResKV struct {
	Key string
	Val string
}

func NewClient(config ConfigEtcdAddr) *EClient {
	var c EClient
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{config.EtcdAddr},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Println(err)
		return &c
	}

	c.cli = client
	return &c
}

func (ec *EClient) Match(serverName string) (kv []ResKV, err error) {
	res, err := ec.cli.Get(context.TODO(), serverName, clientv3.WithPrefix())
	if err != nil {
		return kv, err
	}
	kv = make([]ResKV, 1)
	for _, v := range res.Kvs {
		kv = append(kv, ResKV{
			Key: string(v.Key),
			Val: string(v.Value),
		})
	}
	return
}

func (ec *EClient) MatchAServer(serverName string) (kv ResKV, err error) {
	res, err := ec.cli.Get(context.TODO(), serverName, clientv3.WithPrefix())
	if err != nil {
		return kv, err
	}
	return ResKV{
		Key: string(res.Kvs[0].Key),
		Val: string(res.Kvs[0].Value),
	}, nil
}
