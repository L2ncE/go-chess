package etcd

import (
	"context"
	"errors"
	"fmt"
	client "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"log"
	"time"
)

type Config struct {
	EleName   string
	NodeName  string
	Endpoints []string
}

type NodeServer struct {
	cli       *client.Client
	leaseId   client.LeaseID
	aliveChan <-chan *client.LeaseKeepAliveResponse
	session   *concurrency.Session
	el        *concurrency.Election
	c         *Config
}

func NewNodeServer(config *Config) (s *NodeServer, err error) {
	fmt.Println(config)
	newClient, err := client.New(client.Config{
		Endpoints:   config.Endpoints,
		DialTimeout: 5 * time.Second,
	})
	s = &NodeServer{}
	s.c = config
	fmt.Println(s.c.EleName)
	if err != nil {
		return &NodeServer{}, err
	}
	s.cli = newClient
	if err != nil {
		return &NodeServer{}, err
	}
	if err = s.Leader(); err != nil {
		log.Println("s.Leader() err", err)
		return
	}
	return s, nil
}

func (s *NodeServer) NewLease(ttl int64) (err error) {
	lease := client.NewLease(s.cli)
	l, err := lease.Grant(context.TODO(), ttl)
	if err != nil {
		return
	}
	s.leaseId = l.ID
	aliveChan, err := lease.KeepAlive(context.TODO(), s.leaseId)
	if err != nil {
		return
	}
	s.aliveChan = aliveChan
	return nil
}

func (s *NodeServer) Leader() (err error) {
	session, err := concurrency.NewSession(s.cli, concurrency.WithTTL(5))
	if err != nil {
		return
	}
	election := concurrency.NewElection(session, s.c.EleName)
	s.el = election
	err = s.beLeader()
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

func (s *NodeServer) beLeader() (err error) {
	err = s.el.Campaign(context.TODO(), s.c.NodeName)
	log.Println(s.c.NodeName + "参与竞选")
	if err != nil {
		log.Println("竞选失败", err)
		return err
	}
	log.Println("竞选成功")
	return
}

func (s *NodeServer) Monitor() (ifLeader bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := s.el.Leader(ctx)
	if err == concurrency.ErrElectionNoLeader {
		log.Println("没有leader")
		return false
	}
	log.Println(res.Kvs)
	return true
}

func (s *NodeServer) StartServer() (err error) {
	time.Sleep(2 * time.Second)
	res, err := s.el.Leader(context.TODO())
	if err != nil {
		log.Println(err)
	}
	if string(res.Kvs[0].Value) == s.c.NodeName {
		fmt.Println(s.c.NodeName)
		return nil
	}
	return errors.New("此节点不是leader")
}
