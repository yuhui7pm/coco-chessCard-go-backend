package main

import (
	"common/config"
	"common/logs"
	"common/metrics"
	"context"
	"flag"
	"fmt"
	"os"
	"user/app"
)

var configFile = flag.String("config", "application.yml", "config fill")

func main() {
	// 1. 加载配置
	flag.Parse()
	config.InitConfig(*configFile)

	// 2. 启动监控
	go func() {
		err := metrics.Server(fmt.Sprintf("0.0.0.0:%d", config.Conf.MetricPort))

		if err != nil {
			panic(err)
		}
	}()

	// 3. 启动GRPC服务
	err := app.Run(context.Background())
	if err != nil {
		logs.Print(err.Error())
		// 非零状态码：通常用来表示程序出现了错误或异常情况。
		// 零状态码：通常用来表示程序正常退出。
		os.Exit(-1)
	}
}
