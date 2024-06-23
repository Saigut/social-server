package main

import (
	"context"
	"time"

	"google.golang.org/grpc"
	pb "social_server/src/gen/grpc" // 修改为您的实际代码路径
	. "social_server/src/utils/log"
)

func main() {
	SetupLogger()

	// 设置服务器地址和端口
	address := "localhost:10080"
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		Log.Error("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGrpcApiClient(conn)

	// 调用 SessRegister
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	r, err := c.UmRegister(ctx, &pb.UmRegisterReq{UserName: "user123", Passphrase: "pass123", Email: "john@example.com"})
	if err != nil {
		Log.Error("could not register: %v", err)
	}
	Log.Info("Registration response: %v", r)

	// 调用 SessUserLogin
	l, err := c.SessUserLogin(ctx, &pb.SessUserLoginReq{Username: "user123", Passphrase: "pass123"})
	if err != nil {
		Log.Error("could not login: %v", err)
	}
	Log.Info("Login response: %v", l)
}
