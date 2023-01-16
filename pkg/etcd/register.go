package etcd

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type Register struct {
	client        *clientv3.Client
	leaseId       clientv3.LeaseID //租约id
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	key           string
	value         string
}

// NewRegister 新建注册对象
func NewRegister(endpoint []string, key, value string, lease int64) (*Register, error) {
	var (
		client *clientv3.Client
		err    error
	)
	client, err = clientv3.New(clientv3.Config{
		Endpoints:   endpoint,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		return nil, err
	}
	register := new(Register)
	register.client = client
	register.key = key
	register.value = value

	//设置租约
	if err = register.putKeyWithLease(lease); err != nil {
		return register, err
	}

	return nil, nil
}

//设置租约
func (r *Register) putKeyWithLease(lease int64) error {
	var (
		err                error
		leaseGrantResponse *clientv3.LeaseGrantResponse
		keepAliveChan      <-chan *clientv3.LeaseKeepAliveResponse
	)

	//申请租约
	leaseGrantResponse, err = r.client.Grant(context.TODO(), lease)
	if err != nil {
		return err
	}

	//注册服务与设置租约
	_, err = r.client.Put(context.TODO(), r.key, r.value, clientv3.WithLease(leaseGrantResponse.ID))
	if err != nil {
		return err
	}

	//自动续费
	keepAliveChan, err = r.client.KeepAlive(context.TODO(), leaseGrantResponse.ID)
	if err != nil {
		return err
	}
	r.keepAliveChan = keepAliveChan
	r.leaseId = leaseGrantResponse.ID

	go func() {
		for {
			select {
			case keepResp := <-r.keepAliveChan: // 自动续租的应答
				if keepResp == nil {
					r.deRegister()
				}
			}
		}
	}()
	return nil
}

func (r *Register) deRegister() error {
	defer r.client.Close()
	_, err := r.client.Revoke(context.TODO(), r.leaseId)
	if err != nil {
		return err
	}

	return nil
}

func (r *Register) Close() {
	if r.client != nil {
		r.client.Close()
	}
}
