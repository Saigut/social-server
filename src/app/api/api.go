package api

import (
	"context"
	"fmt"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"net/http"
	"social_server/src/app/service/core"
	. "social_server/src/gen/grpc"
	. "social_server/src/utils/log"
	"strings"
)

type sessCtxT struct {
	Uid string
}

type ModApi struct {
	aGrpcApiServer *grpcApiServer
}

func NewModApi() *ModApi {
	return &ModApi{
		aGrpcApiServer: NewGrpcApiServer(),
	}
}

type grpcApiServer struct {
	Core *core.Core
	UnimplementedGrpcApiServer
}

func NewGrpcApiServer() *grpcApiServer {
	return &grpcApiServer{
		Core: core.NewCore(),
	}
}

func (p *grpcApiServer) SessUserLogin(ctx context.Context, req *SessUserLoginReq) (*SessUserLoginRes, error) {
	return p.Core.SessUserLogin(req)
}

func (p *grpcApiServer) SessUserLogout(ctx context.Context, req *SessUserLogoutReq) (*SessUserLogoutRes, error) {
	return p.Core.SessUserLogout(req)
}

func (p *grpcApiServer) UmRegister(ctx context.Context, req *UmRegisterReq) (*UmRegisterRes, error) {
	return p.Core.UmRegister(req)
}
func (p *grpcApiServer) UmUnregister(ctx context.Context, req *UmUnregisterReq) (*UmUnregisterRes, error) {
	return p.Core.UmUnregister(req)
}
func (p *grpcApiServer) UmAddFriend(ctx context.Context, req *UmAddFriendReq) (*UmAddFriendRes, error) {
	return p.Core.UmAddFriend(req)
}
func (p *grpcApiServer) UmDelFriend(ctx context.Context, req *UmDelFriendReq) (*UmDelFriendRes, error) {
	return p.Core.UmDelFriend(req)
}
func (p *grpcApiServer) UmGetFriendList(ctx context.Context, req *UmGetFriendListReq) (*UmGetFriendListRes, error) {
	return p.Core.UmGetFriendList(req)
}
func (p *grpcApiServer) ChatGetMsgList(ctx context.Context, req *ChatGetMsgListReq) (*ChatGetMsgListRes, error) {
	return p.Core.ChatGetMsgList(req)
}
func (p *grpcApiServer) ChatGetBoxMsgHist(ctx context.Context, req *ChatGetBoxMsgHistReq) (*ChatGetBoxMsgHistRes, error) {
	return p.Core.ChatGetBoxMsgHist(req)
}
func (p *grpcApiServer) ChatSendMsg(ctx context.Context, req *ChatSendMsgReq) (*ChatSendMsgRes, error) {
	return p.Core.ChatSendMsg(req)
}
func (p *grpcApiServer) ChatCreateGroupConv(ctx context.Context, req *ChatCreateGroupConvReq) (*ChatCreateGroupConvRes, error) {
	return p.Core.ChatCreateGroupConv(req)
}
func (p *grpcApiServer) ChatGetGroupConvList(ctx context.Context, req *ChatGetGroupConvListReq) (*ChatGetGroupConvListRes, error) {
	return p.Core.ChatGetGroupConvList(req)
}

func (p *ModApi) StartRpcServer() (error) {
	// 准备 grpc server
	// var opts []grpc.ServerOption
	grpcServer := grpc.NewServer()
	RegisterGrpcApiServer(grpcServer, p.aGrpcApiServer)

	grpcWebServer := grpcweb.WrapServer(
		grpcServer,
		// Enable CORS
		grpcweb.WithOriginFunc(func(origin string) bool { return true }),
	)

	// 创建 HTTP/2 服务器
	h2s := &http2.Server{}

	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			contentType := r.Header.Get("Content-Type")
			Log.Debug("HTTP method: %s, URL: %s, Ver: %s, Content-Type: %s", r.Method, r.URL, r.Proto, contentType)
			if contentType == "application/grpc" && r.ProtoMajor == 2 {
				Log.Debug("HTTP/2 grpc")
				grpcServer.ServeHTTP(w, r)
			} else if strings.Contains(contentType, "application/grpc-web") {
				Log.Debug("HTTP/1.1 grpc-web")
				grpcWebServer.ServeHTTP(w, r)
			} else {
				Log.Debug("Normal HTTP")
				// 这里处理非 gRPC 的 HTTP 请求
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Hello, this is a HTTP request."))
			}
			Log.Debug("Processed.")
		})

	// 创建一个 HTTP 服务器
	httpServer := &http.Server{
		Addr:    "0.0.0.0:10080",
		Handler: h2c.NewHandler(handler, h2s),
	}

	// 启动服务器
	Log.Info("Serving gRPC and HTTP on port 10080")
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to serve: %v", err)
	}

	Log.Info("Quit.")
	return nil
}
