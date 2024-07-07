package discovery

import (
	"common/config"
	"common/logs"
	"context"
	"encoding/json"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

// Register 将grpc注册到etcd
// 原理 创建一个租约 将grpc服务信息注册到etcd并且绑定租约
// 如果过了租约时间，etcd会删除存储的信息
// 可以实现心跳，完成续租，如果etcd没有则重新注册
type Register struct {
	etcdCli     *clientv3.Client                        //etcd连接
	leaseId     clientv3.LeaseID                        //租约id
	DialTimeout int                                     //超时时间 秒
	ttl         int64                                   //租约时间 秒
	keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse // 心跳channel
	info        Server                                  //注册的服务信息
	closeCh     chan struct{}
}

func NewRegister() *Register {
	return &Register{
		DialTimeout: 3,
	}
}

func (register *Register) Close() {
	register.closeCh <- struct{}{}
}

func (register *Register) Register(conf config.EtcdConf) error {
	// 注册消息
	info := Server{
		Name:    conf.Register.Name,
		Addr:    conf.Register.Addr,
		Weight:  conf.Register.Weight,
		Version: conf.Register.Version,
		Ttl:     conf.Register.Ttl,
	}

	// 建立etcd连接
	var err error

	register.etcdCli, err = clientv3.New(clientv3.Config{
		Endpoints:   conf.Addrs,
		DialTimeout: time.Duration(register.DialTimeout) * time.Second,
	})

	if err != nil {
		return err
	}

	register.info = info

	if err = register.register(); err != nil {
		return err
	}

	register.closeCh = make(chan struct{})
	// 放入协程中，根据心跳的结果，做相应操作。
	go register.watcher()

	return nil
}

func (register *Register) register() error {
	var err error

	// 1. 创建租约
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(register.DialTimeout))
	defer cancel()

	if err = register.createRelease(ctx, register.info.Ttl); err != nil {
		return err
	}

	// 2. 心跳检测
	if register.keepAliveCh, err = register.keepAlive(ctx); err != nil {
		return err
	}

	// 3. 绑定租约
	data, _ := json.Marshal(register.info)
	return register.bindRelease(ctx, register.info.BuildRegisterKey(), string(data))
}

// 1.创建租约
func (register *Register) createRelease(ctx context.Context, ttl int64) error {
	grant, err := register.etcdCli.Grant(ctx, ttl)

	if err != nil {
		logs.Error("create release failed err:%v", err)
		return err
	}

	register.leaseId = grant.ID

	return nil
}

// 2. 心跳检测
func (register *Register) keepAlive(context context.Context) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	keepAliveResponse, err := register.etcdCli.KeepAlive(context, register.leaseId)

	if err != nil {
		logs.Error("bind release failed err:%v", err)
		return keepAliveResponse, err
	}

	return keepAliveResponse, nil
}

// 3.绑定租约
func (register *Register) bindRelease(ctx context.Context, key string, value string) error {
	// put动作
	_, err := register.etcdCli.Put(ctx, key, value, clientv3.WithLease(register.leaseId))

	if err != nil {
		logs.Error("bind release failed err:%v", err)
		return err
	}

	return nil
}

// watcher 续约 新注册 close
func (register *Register) watcher() {
	//租约到期了，检测是否需要自动注册
	ticker := time.NewTicker(time.Duration(register.info.Ttl) * time.Second)

	for {
		select {
		case <-register.closeCh:
			// 用户注销
			if err := register.unregister(); err != nil {
				logs.Error("close and unregister failed err:%v", err)
			}

			// 租约撤销
			if _, err := register.etcdCli.Revoke(context.Background(), register.leaseId); err != nil {
				logs.Error("close and revoke lease failed err:%v", err)
			}

			logs.Info("unregister etdc...")
		case res := <-register.keepAliveCh:
			// TODO：我觉的这个判断逻辑有问题
			if res != nil {
				if err := register.register(); err != nil {
					logs.Error("keep alive register failed err:%v", err)
				}
			}
		case <-ticker.C:
			// 心跳检测
			if register.keepAliveCh == nil {
				if err := register.register(); err != nil {
					logs.Error("ticker register failed err:%v", err)
				}
			}
		}
	}
}

// 注销
func (register *Register) unregister() error {
	_, err := register.etcdCli.Delete(context.Background(), register.info.BuildRegisterKey())
	return err
}
