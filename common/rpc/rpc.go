package rpc

import (
	"common/config"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"user/pb"
)

var (
	UserClient pb.UserServiceClient
)

func Init() {
	// etcd解析器。当grpc连接的时候，进行触发，通过提供的addr地址，去etcd中进行查找
	//找服务的地址
	userDomain := config.Conf.Domain["user"]
	InitClient(userDomain.Name, userDomain.LoadBalance, &UserClient)
}

func InitClient(name string, loadBalance bool, client interface{}) {
	address := fmt.Sprintf("etcd:///%s", name)
	connection, err := grpc.DialContext(context.TODO(), address)
	if err != nil {
		log.Fatalf("rpc connect etcd error:%v", err)
	}

	switch c := client.(type) {
	case *pb.UserServiceClient:
		*c = pb.NewUserServiceClient(connection)
	default:
		log.Fatal("unsupported")
	}
}
