package database

import (
	"common/config"
	"common/logs"
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisManager struct {
	Client        *redis.Client        //单机
	ClusterClient *redis.ClusterClient //集群
}

func NewRedis() *RedisManager {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var client *redis.Client               //单机
	var clusterClient *redis.ClusterClient //集群

	addrs := config.Conf.Database.RedisConf.ClusterAddrs

	if len(addrs) == 0 {
		// 非集群，单节点
		client = redis.NewClient(&redis.Options{
			Addr:         config.Conf.Database.RedisConf.Addr,
			PoolSize:     config.Conf.Database.RedisConf.PoolSize,
			MinIdleConns: config.Conf.Database.RedisConf.MinIdleConns,
			Password:     config.Conf.Database.RedisConf.Password,
		})
	} else {
		clusterClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        config.Conf.Database.RedisConf.ClusterAddrs,
			PoolSize:     config.Conf.Database.RedisConf.PoolSize,
			MinIdleConns: config.Conf.Database.RedisConf.MinIdleConns,
			Password:     config.Conf.Database.RedisConf.Password,
		})
	}

	if client != nil {
		if err := client.Ping(ctx); err != nil {
			logs.Fatal("redis ping err:%v", err)
			return nil
		}
	}

	if clusterClient != nil {
		if err := clusterClient.Ping(ctx); err != nil {
			logs.Fatal("redis client ping err:%v", err)
			return nil
		}
	}

	return &RedisManager{
		Client:        client,
		ClusterClient: clusterClient,
	}
}

func (redis *RedisManager) Close() {
	if redis.Client != nil {
		err := redis.Client.Close()

		if err != nil {
			logs.Error("close redis err:%v", err)
		}
	}

	if redis.ClusterClient != nil {
		err := redis.ClusterClient.Close()

		if err != nil {
			logs.Error("close redis cluster err:%v", err)
		}
	}
}

func (redis *RedisManager) Set(ctx context.Context, key, value string, expireTime time.Duration) error {
	if redis.Client != nil {
		return redis.Client.Set(ctx, key, value, expireTime).Err()
	}

	if redis.ClusterClient != nil {
		return redis.ClusterClient.Set(ctx, key, value, expireTime).Err()
	}

	return nil
}
