package app

import (
	"common/config"
	"common/discovery"
	"common/logs"
	"context"
	"core/repo"
	"fmt"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"user/internal/service"
	"user/pb"
)

// 启动程序 grpc http服务 启用日志 启用数据库
func Run(ctx context.Context) error {
	// 1. 日志库：info fatal error debug
	logs.InitLog(config.Conf.AppName)
	// 2. etdc注册中心  grpc服务注册到etdc中 客户端访问的时候 通过etdc获取grpc的·地址
	register := discovery.NewRegister()

	server := grpc.NewServer()
	manger := repo.New()

	// 使用协程，避免阻塞进程
	go func() {
		lis, err := net.Listen("tcp", config.Conf.Grpc.Addr)
		if err != nil {
			logs.Fatal("user grpc server listen err:%v", err)
		}

		// 注册grpc service，需要数据库mongo 和 redis
		err = register.Register(config.Conf.Etcd)
		if err != nil {
			logs.Fatal("user grpc server register etdc err:%v", err)
		}

		// 初始化数据库操作
		pb.RegisterUserServiceServer(server, service.NewAccountService(manger))

		// 阻塞进程
		err = server.Serve(lis)
		if err != nil {
			logs.Fatal("user grpc server failed err:%v", err)
		}
	}()

	stop := func() {
		server.Stop()
		register.Close()
		manger.Close()

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
