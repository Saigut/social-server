package api

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"math/rand"
	"net/http"
	. "social_server/src/app/service/controller"
	"social_server/src/app/service/sess_mgmt"
	. "social_server/src/app/service/user_mgmt"
	. "social_server/src/gen/grpc"
	. "social_server/src/utils/log"
	"strings"
	"time"
)

/*

session:
	- user_login
	- user_logout

api:

* user-mgmt
	- register
	- unregister
	- login
	- logout
	- add_friends
	- del_friends
	- list_friends

* chat
	- get_chat_msg
	- get_chat_msg_hist_with
	- send_chat_msg_to
*/

type sessCtxT struct {
	Uid string
}

type ModApiT struct {
	sessMap map[string]sessCtxT

	Usermgmt *UserMgmtT
	Controller *ControllerT

	grpcApiServer *grpcApiServerT
}

type grpcApiServerT struct {
	ModApi *ModApiT
	UnimplementedGrpcApiServer
}

func (p *grpcApiServerT) generateToken(uid string) (token string, err error) {
	rand.Seed(time.Now().UnixNano())
	originalStr := fmt.Sprintf("%s%d%d", uid, time.Now().Unix(), rand.Intn(10000))

	// 计算 MD5 哈希值
	hash := md5.Sum([]byte(originalStr))
	token = hex.EncodeToString(hash[:])

	_, isExisted := p.ModApi.sessMap[token]
	if isExisted {
		errStr := "token existed."
		Log.Error(errStr)
		return "", status.Errorf(codes.Internal, errStr)
	}

	return token, nil
}

func (p *grpcApiServerT) SessMapInsert(sessToken string, sessCtx sessCtxT) (error) {
	_, isExisted := p.ModApi.sessMap[sessToken]
	if isExisted {
		p.ModApi.sessMap[sessToken] = sessCtx
		return nil
	}
	p.ModApi.sessMap[sessToken] = sessCtx
	return nil
}

func (p *grpcApiServerT) SessMapDel(sessToken string) (error) {
	delete(p.ModApi.sessMap, sessToken)
	return nil
}

func (p *grpcApiServerT) SessMapGet(sessToken string) (sessCtxT, error) {
	ctx, isExisted := p.ModApi.sessMap[sessToken]
	if !isExisted {
		return sessCtxT{}, status.Errorf(codes.Unimplemented, "method SessRegister not found")
	}
	return ctx, nil
}

func (p *grpcApiServerT) SessUserLogin(ctx context.Context, req *SessUserLoginReq) (*SessUserLoginRes, error) {
	var res SessUserLoginRes

	// 1. uid 和 passphase
	var loginParam UmLoginParam
	loginParam.Uid = req.GetUserid()
	loginParam.Passwd = req.GetPassphase()

	// 2. 检查 uid 和 passphase
	//		error: 返回失败信息
	err := p.ModApi.Usermgmt.Login(&loginParam)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	// 3. 创建 sess 信息
	var token string
	token, err = p.generateToken(req.Userid)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	// 4. 保存 sess 信息
	var sessCtx sessCtxT
	sessCtx.Uid = req.Userid
	err = p.SessMapInsert(token, sessCtx)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	// 5. 返回包括 sess token 在内的成功信息
	res.Ret = 0
	res.SessToken = token
	return &res, nil
}

func (p *grpcApiServerT) SessUserLogout(ctx context.Context, req *SessUserLogoutReq) (*SessUserLogoutRes, error) {
	sessId, err := p.getSessId(ctx)
	if err != nil {
		return nil, err
	}
	return p.ModApi.Controller.SessUserLogout(sessId, req)
}

func (p *grpcApiServerT) UmRegister(ctx context.Context, req *UmRegisterReq) (*UmRegisterRes, error) {
	return p.ModApi.Controller.UmRegister(req)
}
func (p *grpcApiServerT) UmUnregister(ctx context.Context, req *UmUnregisterReq) (*UmUnregisterRes, error) {
	sessId, err := p.getSessId(ctx)
	if err != nil {
		return nil, err
	}
	return p.ModApi.Controller.UmUnregister(sessId, req)
}
func (p *grpcApiServerT) UmAddFriends(ctx context.Context, req *UmAddFriendsReq) (*UmAddFriendsRes, error) {
	sessId, err := p.getSessId(ctx)
	if err != nil {
		return nil, err
	}
	return p.ModApi.Controller.UmAddFriends(sessId, req)
}
func (p *grpcApiServerT) UmDelFriends(ctx context.Context, req *UmDelFriendsReq) (*UmDelFriendsRes, error) {
	sessId, err := p.getSessId(ctx)
	if err != nil {
		return nil, err
	}
	return p.ModApi.Controller.UmDelFriends(sessId, req)
}
func (p *grpcApiServerT) UmListFriends(ctx context.Context, req *UmListFriendsReq) (*UmListFriendsRes, error) {
	sessId, err := p.getSessId(ctx)
	if err != nil {
		return nil, err
	}
	return p.ModApi.Controller.UmListFriends(sessId, req)
}
func (p *grpcApiServerT) ChatGetChatMsg(ctx context.Context, req *ChatGetChatMsgReq) (*ChatGetChatMsgRes, error) {
	sessId, err := p.getSessId(ctx)
	if err != nil {
		return nil, err
	}
	return p.ModApi.Controller.ChatGetChatMsg(sessId, req)
}
func (p *grpcApiServerT) ChatGetChatMsgHistWith(ctx context.Context, req *ChatGetChatMsgHistWithReq) (*ChatGetChatMsgHistWithRes, error) {
	sessId, err := p.getSessId(ctx)
	if err != nil {
		return nil, err
	}
	return p.ModApi.Controller.ChatGetChatMsgHistWith(sessId, req)
}
func (p *grpcApiServerT) ChatSendChatMsgTo(ctx context.Context, req *ChatSendChatMsgToReq) (*ChatSendChatMsgToRes, error) {
	sessId, err := p.getSessId(ctx)
	if err != nil {
		return nil, err
	}
	return p.ModApi.Controller.ChatSendChatMsgTo(sessId, req)
}
func (p *grpcApiServerT) ChatGetChatConvId(ctx context.Context, req *ChatGetChatConvIdReq) (*ChatGetChatConvIdRes, error) {
	sessId, err := p.getSessId(ctx)
	if err != nil {
		return nil, err
	}
	return p.ModApi.Controller.ChatGetChatConvId(sessId, req)
}

func (p *grpcApiServerT) getSessId(ctx context.Context) (sess_mgmt.SessIdT, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "missing session info")
	}
	sessId, ok := md["session-id"]
	if !ok || len(sessId) == 0 {
		return "", status.Errorf(codes.Unauthenticated, "missing session id")
	}
	return sess_mgmt.SessIdT(sessId[0]), nil
}

// validateSessionToken 验证会话标识的有效性
func (p *ModApiT) validateSessionToken(token string) bool {
	// 这里简化处理，直接比较会话标识
	// 实际应用中，你可能需要查询数据库或会话存储来验证会话标识
	_, err := p.grpcApiServer.SessMapGet(token)
	if err != nil {
		return false
	} else {
		return true
	}
}

func (p *ModApiT) StartRpcServer() (error) {
	// 准备 grpc server
	//var opts []grpc.ServerOption
	grpcServer := grpc.NewServer()
	p.grpcApiServer = new(grpcApiServerT)
	/// Fixme
	p.grpcApiServer.ModApi = p
	RegisterGrpcApiServer(grpcServer, p.grpcApiServer)

	// 创建 HTTP/2 服务器
	h2s := &http2.Server{}

	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			//Log.Info("Debug: HTTP method: %s, URL: %s, Header: %v", r.Method, r.URL, r.Header)
			Log.Info("Debug: HTTP method: %s, URL: %s", r.Method, r.URL)
			if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
				grpcServer.ServeHTTP(w, r)
			} else {
				// 这里处理非 gRPC 的 HTTP 请求
				// 例如: w.Write([]byte("Hello, this is a HTTP request."))
			}
		})

	// 创建一个 HTTP 服务器
	httpServer := &http.Server{
		Addr: "0.0.0.0:10080",
		Handler: h2c.NewHandler(handler, h2s),
	}

	// 启动服务器
	Log.Info("Serving gRPC and HTTP on port 10080")
	if err := httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}
