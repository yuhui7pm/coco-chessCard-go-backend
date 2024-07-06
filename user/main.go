package main

import (
	"common/config"
	"common/metrics"
	"flag"
	"fmt"
)

var configFile = flag.String("config", "application.yml", "config fill")

func main() {
	// 1. 加载配置
	flag.Parse()
	config.InitConfig(*configFile)

	// 2. 启动监控
	go func() {
		err := metrics.Server(fmt.Sprintf("0.0.0.0:%d", config.Conf.MetricPort))
		fmt.Println("1111")

		if err != nil {
			panic(err)
		}
	}()

	// 3. 启动GRPC服务
	select {}
}
