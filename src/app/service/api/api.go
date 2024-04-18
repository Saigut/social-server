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
	"log"
	"math/rand"
	"net/http"
	. "social_server/src/gen/grpc"
	. "social_server/src/app/service/user_mgmt"
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

* post
	stream
	- put_post
	- get_video_hls
	msg
	- get_post_list
	- get_post_metadata
	- get_explorer_post_list
	- get_likes
	- do_like
	- undo_like
	- get_comments
	- add_comment
	- del_comment
*/

type sessCtxT struct {
	Uid string
}

type ModApiT struct {
	grpcApiServer *grpcApiServerT
}

type grpcApiServerT struct {
	ModApi *ModApiT
	sessMap map[string]sessCtxT
	Usermgmt *UsermgmtT
	UnimplementedGrpcApiServer
}

func (p *grpcApiServerT) generateToken(uid string) (token string, err error) {
	rand.Seed(time.Now().UnixNano())
	originalStr := fmt.Sprintf("%s%d%d", uid, time.Now().Unix(), rand.Intn(10000))

	// 计算 MD5 哈希值
	hash := md5.Sum([]byte(originalStr))
	token = hex.EncodeToString(hash[:])

	_, isExisted := p.sessMap[token]
	if isExisted {
		errStr := "token existed."
		log.Fatal(errStr)
		return "", status.Errorf(codes.Internal, errStr)
	}

	return token, nil
}

func (p *grpcApiServerT) SessMapInsert(sessToken string, sessCtx sessCtxT) (error) {
	ctx, isExisted := p.sessMap[sessToken]
	if isExisted {
		ctx = sessCtx
		return nil
	}
	p.sessMap[sessToken] = sessCtx
	return nil
}

func (p *grpcApiServerT) SessMapDel(sessToken string) (error) {
	delete(p.sessMap, sessToken)
	return nil
}

func (p *grpcApiServerT) SessMapGet(sessToken string) (sessCtxT, error) {
	ctx, isExisted := p.sessMap[sessToken]
	if !isExisted {
		return sessCtxT{}, status.Errorf(codes.Unimplemented, "method SessRegister not found")
	}
	return ctx, nil
}

func (p *grpcApiServerT) SessRegister(ctx context.Context, req *SessRegisterReq) (*SessRegisterRes, error) {
	// 1. 获取用 uid
	return nil, status.Errorf(codes.Unimplemented, "method SessRegister not implemented")
}
func (p *grpcApiServerT) SessUnregister(ctx context.Context, req *SessUnregisterReq) (*SessUnregisterRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SessUnregister not implemented")
}
func (p *grpcApiServerT) SessUserLogin(ctx context.Context, req *SessUserLoginReq) (*SessUserLoginRes, err error) {
	var res SessUserLoginRes

	// 1. uid 和 passphase
	var loginParam UmLoginParam
	loginParam.Uid = req.Userid
	loginParam.Passwd = req.Passwd

	// 2. 检查 uid 和 passphase
	//		error: 返回失败信息
	err = p.Usermgmt.Login(&loginParam)
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
	return nil, status.Errorf(codes.Unimplemented, "method SessUserLogout not implemented")
}
func (p *grpcApiServerT) UmAddFriends(ctx context.Context, req *UmAddFriendsReq) (*UmAddFriendsRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UmAddFriends not implemented")
}
func (p *grpcApiServerT) UmDelFriends(ctx context.Context, req *UmDelFriendsReq) (*UmDelFriendsRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UmDelFriends not implemented")
}
func (p *grpcApiServerT) UmListFriends(ctx context.Context, req *UmListFriendsReq) (*UmListFriendsRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UmListFriends not implemented")
}
func (p *grpcApiServerT) ChatGetChatMsg(ctx context.Context, req *ChatGetChatMsgReq) (*ChatGetChatMsgRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChatGetChatMsg not implemented")
}
func (p *grpcApiServerT) ChatGetChatMsgHistWith(ctx context.Context, req *ChatGetChatMsgHistWithReq) (*ChatGetChatMsgHistWithRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChatGetChatMsgHistWith not implemented")
}
func (p *grpcApiServerT) ChatSendChatMsgTo(ctx context.Context, req *ChatSendChatMsgToReq) (*ChatSendChatMsgToRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChatSendChatMsgTo not implemented")
}
func (p *grpcApiServerT) ChatGetChatConvId(ctx context.Context, req *ChatGetChatConvIdReq) (*ChatGetChatConvIdRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChatGetChatConvId not implemented")
}
func (p *grpcApiServerT) PostPutPost(ctx context.Context, req *PostPutPostReq) (*PostPutPostRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostPutPost not implemented")
}
func (p *grpcApiServerT) PostGetVideoHls(ctx context.Context, req *PostGetVideoHlsReq) (*PostGetVideoHlsRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostGetVideoHls not implemented")
}
func (p *grpcApiServerT) PostGetPostList(ctx context.Context, req *PostGetPostListReq) (*PostGetPostListRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostGetPostList not implemented")
}
func (p *grpcApiServerT) PostGetPostMetadata(ctx context.Context, req *PostGetPostMetadataReq) (*PostGetPostMetadataRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostGetPostMetadata not implemented")
}
func (p *grpcApiServerT) PostGetExplorerVideoList(ctx context.Context, req *PostGetExplorerVideoListReq) (*PostGetExplorerVideoListRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostGetExplorerVideoList not implemented")
}
func (p *grpcApiServerT) PostGetLikes(ctx context.Context, req *PostGetLikesReq) (*PostGetLikesRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostGetLikes not implemented")
}
func (p *grpcApiServerT) PostDoLike(ctx context.Context, req *PostDoLikeReq) (*PostDoLikeRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostDoLike not implemented")
}
func (p *grpcApiServerT) PostUndoLike(ctx context.Context, req *PostUndoLikeReq) (*PostUndoLikeRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostUndoLike not implemented")
}
func (p *grpcApiServerT) PostGetComments(ctx context.Context, req *PostGetCommentsReq) (*PostGetCommentsRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostGetComments not implemented")
}
func (p *grpcApiServerT) PostAddComment(ctx context.Context, req *PostAddCommentReq) (*PostAddCommentRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostAddComment not implemented")
}
func (p *grpcApiServerT) PostDelComment(ctx context.Context, req *PostDelCommentReq) (*PostDelCommentRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostDelComment not implemented")
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

// sessionAuthInterceptor 是一个服务端拦截器，用于验证会话标识
func (p *ModApiT) sessionAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 允许 Login 方法不经过会话验证
	if info.FullMethod != "/gen_grpc.GrpcApi/SessRegister" && info.FullMethod != "/gen_grpc.GrpcApi/SessUserLogin" {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			// 提取会话标识
			values := md["session-token"]
			if len(values) == 0 || !p.validateSessionToken(values[0]) {
				// 会话验证失败
				return nil, status.Errorf(codes.Unauthenticated, "invalid session token")
			}
		} else {
			return nil, status.Errorf(codes.Unauthenticated, "missing session token")
		}
	}
	// 继续处理请求
	return handler(ctx, req)
}

func (p *ModApiT) StartRpcServer() (error) {
	// 准备 grpc server
	//var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(p.sessionAuthInterceptor))
	p.grpcApiServer = new(grpcApiServerT)
	/// Fixme
	p.grpcApiServer.ModApi = p
	RegisterGrpcApiServer(grpcServer, p.grpcApiServer)

	// 创建 HTTP/2 服务器
	h2s := &http2.Server{}

	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
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
	fmt.Println("Serving gRPC and HTTP on port 80")
	if err := httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}


	return nil
}
