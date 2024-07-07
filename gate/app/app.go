package app

import (
	"common/config"
	"common/logs"
	"context"
	"fmt"
	"gate/router"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 启动程序 grpc http服务 启用日志 启用数据库
func Run(ctx context.Context) error {
	// 1. 日志库：info fatal error debug
	logs.InitLog(config.Conf.AppName)
	// 使用协程，避免阻塞进程
	go func() {
		// gin启动  注册路由
		routerApi := router.RegisterRouter()

		// http接口
		if err := routerApi.Run(fmt.Sprintf(":%s", config.Conf.HttpPort)); err != nil {
			logs.Fatal("gate gin run err:%v", err)
		}
	}()

	stop := func() {
		time.Sleep(3 * time.Second)
		fmt.Println("the func to stop app")
	}
	// 期望有一个优雅启停，遇到中断信号，终止信号，期望直接退出
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGHUP)
	for {
		// select 语句用于等待多个通道操作。当其中一个通道操作可以进行时，
		// select 语句就会执行相应的 case 语句。它主要用于处理并发操作，特别是通道的读写。
		select {
		case <-ctx.Done():
			stop()
			return nil
		case signalTemp := <-channel:
			switch signalTemp {
			case syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT:
				stop()
				logs.Print("user app quit it self")
				return nil
			case syscall.SIGHUP:
				stop()
				logs.Print("hangup user app quit it self")
				return nil
			default:
				return nil
			}
		}
	}
}
