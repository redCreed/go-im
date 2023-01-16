package etcd

import (
	"context"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"sync"
	"time"
)

type Discovery struct {
	client  *clientv3.Client
	srvList map[string]string
	lock    sync.Mutex
}

// NewDiscovery 发现服务
func NewDiscovery(endpoints []string) (*Discovery, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return &Discovery{
		client:  cli,
		srvList: make(map[string]string),
	}, nil
}

// WatchSrv  初始化服务列表和监视
func (s *Discovery) WatchSrv(prefix string) error {
	//根据前缀获取现有的key
	resp, err := s.client.Get(context.TODO(), prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	for _, ev := range resp.Kvs {
		s.SetSrvList(string(ev.Key), string(ev.Value))
	}

	//监视前缀，修改变更的server
	go s.watcher(prefix)
	return nil
}

//watcher 监听前缀
func (s *Discovery) watcher(prefix string) {
	rch := s.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT: //修改或者新增
				s.SetSrvList(string(ev.Kv.Key), string(ev.Kv.Value))
			case mvccpb.DELETE: //删除
				s.DelSrvList(string(ev.Kv.Key))
			}
		}
	}
}

// SetSrvList  新增服务地址
func (s *Discovery) SetSrvList(key, val string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.srvList[key] = val
}

// DelSrvList 删除服务地址
func (s *Discovery) DelSrvList(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.srvList, key)
}

// GetSrv 获取服务地址
func (s *Discovery) GetSrv() []string {
	s.lock.Lock()
	defer s.lock.Unlock()
	addrs := make([]string, 0)

	for _, v := range s.srvList {
		addrs = append(addrs, v)
	}
	return addrs
}

//Close 关闭服务
func (s *Discovery) Close() error {
	return s.client.Close()
}
