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
func (p *grpcApiServer) UmUserUpdateInfo(ctx context.Context, req *UmUserUpdateInfoReq) (*UmUserUpdateInfoRes, error) {
	return p.Core.UmUserUpdateInfo(req)
}
func (p *grpcApiServer) UmContactGetList(ctx context.Context, req *UmContactGetListReq) (*UmContactGetListRes, error) {
	return p.Core.UmContactGetList(req)
}
func (p *grpcApiServer) UmContactGetInfo(ctx context.Context, req *UmContactGetInfoReq) (*UmContactGetInfoRes, error) {
	return p.Core.UmContactGetInfo(req)
}
func (p *grpcApiServer) UmContactFind(ctx context.Context, req *UmContactFindReq) (*UmContactFindRes, error) {
	return p.Core.UmContactFind(req)
}
func (p *grpcApiServer) UmContactAddRequest(ctx context.Context, req *UmContactAddRequestReq) (*UmContactAddRequestRes, error) {
	return p.Core.UmContactAddRequest(req)
}
func (p *grpcApiServer) UmContactAccept(ctx context.Context, req *UmContactAcceptReq) (*UmContactAcceptRes, error) {
	return p.Core.UmContactAccept(req)
}
func (p *grpcApiServer) UmContactReject(ctx context.Context, req *UmContactRejectReq) (*UmContactRejectRes, error) {
	return p.Core.UmContactReject(req)
}
func (p *grpcApiServer) UmContactDel(ctx context.Context, req *UmContactDelReq) (*UmContactDelRes, error) {
	return p.Core.UmContactDel(req)
}
func (p *grpcApiServer) UmGroupGetList(ctx context.Context, req *UmGroupGetListReq) (*UmGroupGetListRes, error) {
	return p.Core.UmGroupGetList(req)
}
func (p *grpcApiServer) UmGroupGetInfo(ctx context.Context, req *UmGroupGetInfoReq) (*UmGroupGetInfoRes, error) {
	return p.Core.UmGroupGetInfo(req)
}
func (p *grpcApiServer) UmGroupUpdateInfo(ctx context.Context, req *UmGroupUpdateInfoReq) (*UmGroupUpdateInfoRes, error) {
	return p.Core.UmGroupUpdateInfo(req)
}
func (p *grpcApiServer) UmGroupFind(ctx context.Context, req *UmGroupFindReq) (*UmGroupFindRes, error) {
	return p.Core.UmGroupFind(req)
}
func (p *grpcApiServer) UmGroupCreate(ctx context.Context, req *UmGroupCreateReq) (*UmGroupCreateRes, error) {
	return p.Core.UmGroupCreate(req)
}
func (p *grpcApiServer) UmGroupDelete(ctx context.Context, req *UmGroupDeleteReq) (*UmGroupDeleteRes, error) {
	return p.Core.UmGroupDelete(req)
}
func (p *grpcApiServer) UmGroupGetMemList(ctx context.Context, req *UmGroupGetMemListReq) (*UmGroupGetMemListRes, error) {
	return p.Core.UmGroupGetMemList(req)
}
func (p *grpcApiServer) UmGroupJoinRequest(ctx context.Context, req *UmGroupJoinRequestReq) (*UmGroupJoinRequestRes, error) {
	return p.Core.UmGroupJoinRequest(req)
}
func (p *grpcApiServer) UmGroupAccept(ctx context.Context, req *UmGroupAcceptReq) (*UmGroupAcceptRes, error) {
	return p.Core.UmGroupAccept(req)
}
func (p *grpcApiServer) UmGroupReject(ctx context.Context, req *UmGroupRejectReq) (*UmGroupRejectRes, error) {
	return p.Core.UmGroupReject(req)
}
func (p *grpcApiServer) UmGroupLeave(ctx context.Context, req *UmGroupLeaveReq) (*UmGroupLeaveRes, error) {
	return p.Core.UmGroupLeave(req)
}
func (p *grpcApiServer) UmGroupAddMem(ctx context.Context, req *UmGroupAddMemReq) (*UmGroupAddMemRes, error) {
	return p.Core.UmGroupAddMem(req)
}
func (p *grpcApiServer) UmGroupDelMem(ctx context.Context, req *UmGroupDelMemReq) (*UmGroupDelMemRes, error) {
	return p.Core.UmGroupDelMem(req)
}
func (p *grpcApiServer) UmGroupUpdateMem(ctx context.Context, req *UmGroupUpdateMemReq) (*UmGroupUpdateMemRes, error) {
	return p.Core.UmGroupUpdateMem(req)
}

func (p *grpcApiServer) ChatSendMsg(ctx context.Context, req *ChatSendMsgReq) (*ChatSendMsgRes, error) {
	return p.Core.ChatSendMsg(req)
}
func (p *grpcApiServer) ChatMarkRead(ctx context.Context, req *ChatMarkReadReq) (*ChatMarkReadRes, error) {
	return p.Core.ChatMarkRead(req)
}

func (p *grpcApiServer) GetUpdateList(ctx context.Context, req *GetUpdateListReq) (*GetUpdateListRes, error) {
	return p.Core.GetUpdateList(req)
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
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Hello, this is a HTTP request."))
			}
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
