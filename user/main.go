package main

import (
	"common/config"
	"flag"
)

var configFile = flag.String("config", "application.yml", "config fill")

func main() {
	// 1. 加载配置
	flag.Parse()
	config.InitConfig(*configFile)

	// 2. 启动监控

	// 3. 启动GRPC服务
}
