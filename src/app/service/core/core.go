package core

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"social_server/src/app/common/types"
	"social_server/src/app/data"
	. "social_server/src/app/service/chat"
	"social_server/src/app/service/sess_mgmt"
	"social_server/src/app/service/user_mgmt"
	"social_server/src/gen/grpc"
	. "social_server/src/utils/log"
	"time"
)

type Core struct {
	userMgmt *user_mgmt.UserMgmt
	sessMgmt *sess_mgmt.SessMgmt
	chat     *Chat
}

func NewCore() *Core {
	// 获取 ENV_PATH 环境变量的值
	envPath := os.Getenv("ENV_PATH")

	// 检查 ENV_PATH 是否被设置
	if envPath != "" {
		// 加载指定路径的 .env 文件
		err := godotenv.Load(envPath)
		if err != nil {
			log.Fatalf("Error loading .env file from %s", envPath)
		}
	} else {
		// 加载默认路径的 .env 文件
		err := godotenv.Load()
		if err != nil {
			log.Println("Error loading default .env file")
		}
	}

	storage := data.NewStorage()
	storage.Init()
	cache := data.NewCache()

	p := &Core{
		userMgmt: user_mgmt.NewUserMgmt(storage, cache),
		sessMgmt: sess_mgmt.NewSessMgmt(storage, cache),
		chat:     NewChat(storage, cache),
	}

	return p
}

func (p *Core) SessUserLogin(req *gen_grpc.SessUserLoginReq) (*gen_grpc.SessUserLoginRes, error) {
	var err error
	var res gen_grpc.SessUserLoginRes

	// 校验用户
	var uaParam types.UmUserAuthenticateParam
	uaParam.Username = req.GetUsername()
	uaParam.Passphase = req.GetPassword()
	var pass bool
	pass, err = p.userMgmt.UserAuthenticate(&uaParam)
	if err != nil {
		Log.Error("SessUserLogin: UserAuthenticate: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}
	if !pass {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UserFailedToAuth
		return &res, nil
	}

	// 检查已登录会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtxByUsername(req.GetUsername())
	if err == nil {
		res.SessId = string(sessCtx.SessId)
		Log.Debug("SessUserLogin: SessId: %s", res.SessId)
		res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
		return &res, nil
	}

	// 若未登录，创建新会话
	var sessId types.SessId
	sessId, err = p.sessMgmt.CreateSess(req.GetUsername(), 2*3600)
	if err != nil {
		Log.Error("SessUserLogin: CreateSess: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}
	res.SessId = string(sessId)
	Log.Debug("SessUserLogin: SessId: %s", res.SessId)
	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) SessUserLogout(req *gen_grpc.SessUserLogoutReq) (*gen_grpc.SessUserLogoutRes, error) {
	var err error
	var res gen_grpc.SessUserLogoutRes

	// 获取会话
	_, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("SessUserLogout: GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 销毁会话
	err = p.sessMgmt.DeleteSess(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("SessUserLogout: DeleteSess: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmRegister(req *gen_grpc.UmRegisterReq) (*gen_grpc.UmRegisterRes, error) {
	var err error
	var res gen_grpc.UmRegisterRes

	// 用户名、密码、邮箱的合规性检查
	var validParam types.UmUserInfoValidateParam
	validParam.Username = req.GetUsername()
	validParam.Passphase = req.GetPassword()
	validParam.Email = req.GetEmail()
	err = p.userMgmt.UserInfoValidate(&validParam)
	if err != nil {
		Log.Error("UmRegister: UserInfoValidate: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 检查用户名是否已存在
	var isExist bool
	isExist, err = p.userMgmt.IsUsernameExisted(req.GetUsername())
	if err != nil {
		Log.Error("UmRegister: IsUsernameExisted: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}
	if isExist {
		Log.Error("UmRegister: User already registered", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UserAlreadyRegistered
		return &res, nil
	}

	// 创建用户
	var regParam types.UmRegisterParam
	regParam.Username = req.GetUsername()
	regParam.Passwd = req.GetPassword()
	regParam.Email = req.GetEmail()
	err = p.userMgmt.Register(&regParam)
	if err != nil {
		Log.Error("UmRegister: Register: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmUnregister(req *gen_grpc.UmUnregisterReq) (*gen_grpc.UmUnregisterRes, error) {
	var err error
	var res gen_grpc.UmUnregisterRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("UmUnregister: GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 注销用户
	var unRegParam types.UmUnregisterParam
	unRegParam.Uid = sessCtx.Uid
	err = p.userMgmt.Unregister(&unRegParam)
	if err != nil {
		Log.Error("UmUnregister: Unregister: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 销毁会话
	err = p.sessMgmt.DeleteSess(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("UmUnregister: DeleteSess: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmAddFriend(req *gen_grpc.UmAddFriendReq) (*gen_grpc.UmAddFriendRes, error) {
	var err error
	var res gen_grpc.UmAddFriendRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("UmAddFriend: GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 添加好友
	var addFriendParam types.UmAddFriendsParam
	addFriendParam.Uid = sessCtx.Uid
	addFriendParam.FriendUid = req.GetFriend().GetUid()
	err = p.userMgmt.AddFriends(&addFriendParam)
	if err != nil {
		Log.Error("UmAddFriend: AddFriends: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmDelFriend(req *gen_grpc.UmDelFriendReq) (*gen_grpc.UmDelFriendRes, error) {
	var err error
	var res gen_grpc.UmDelFriendRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("UmDelFriend: GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 删除好友
	var delFriendParam types.UmDelFriendsParam
	delFriendParam.Uid = sessCtx.Uid
	delFriendParam.FriendUid = req.GetFriendUid()
	err = p.userMgmt.DelFriend(&delFriendParam)
	if err != nil {
		Log.Error("UmDelFriend: DelFriend: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmGetFriendList(req *gen_grpc.UmGetFriendListReq) (*gen_grpc.UmGetFriendListRes, error) {
	var err error
	var res gen_grpc.UmGetFriendListRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("UmGetFriendList: GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 获取好友列表
	var friendUidList []uint64
	friendUidList, err = p.userMgmt.GetFriendList(sessCtx.Uid)
	if err != nil {
		Log.Error("UmGetFriendList: GetFriendList: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	for _, uid := range friendUidList {
		var friendInfo gen_grpc.UmFriendInfo
		friendInfo.Uid = uid
		friendInfo.Nickname = ""
		friendInfo.NoteName = ""
		res.FriendList = append(res.FriendList, &friendInfo)
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) ChatGetMsgList(req *gen_grpc.ChatGetMsgListReq) (*gen_grpc.ChatGetMsgListRes, error) {
	var err error
	var res gen_grpc.ChatGetMsgListRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("ChatGetMsgList: GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 获取聊天消息
	var msgList []types.ChatMsgOfConv
	msgList, err = p.chat.GetChatMsgList(sessCtx.Uid, req.GetLocalSeqId())
	if err != nil {
		Log.Error("ChatGetMsgList: GetChatMsgList: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}
	for _, aConvMsg := range msgList {
		aBoxMsgApi := &gen_grpc.ChatBoxMsg{
			PeerId: &gen_grpc.ChatPeerId{},
			Msg:    &gen_grpc.ChatMsg{},
			SeqId:  aConvMsg.SeqId,
		}
		if aConvMsg.PeerId.PeerIdType == types.EmPeerIdType_PeerUid {
			aBoxMsgApi.PeerId.PeerIdUnion = &gen_grpc.ChatPeerId_PeerUid{PeerUid: aConvMsg.PeerId.PeerUid}
		} else {
			aBoxMsgApi.PeerId.PeerIdUnion = &gen_grpc.ChatPeerId_GroupConvId{GroupConvId: aConvMsg.PeerId.GroupConvId}
		}
		aBoxMsgApi.Msg.MsgType = gen_grpc.ChatMsgType(aConvMsg.Msg.MsgType)
		aBoxMsgApi.Msg.SentTsMs = aConvMsg.Msg.SentTsMs
		aBoxMsgApi.Msg.SenderUid = aConvMsg.Msg.SenderUid
		aBoxMsgApi.Msg.MsgContent = aConvMsg.Msg.MsgContent
		Log.Debug("aConvMsg.Msg.MsgContent: %s", aConvMsg.Msg.MsgContent)
		Log.Debug("aBoxMsgApi.Msg.MsgContent: %s", aBoxMsgApi.Msg.MsgContent)
		res.MsgList = append(res.MsgList, aBoxMsgApi)
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) ChatGetBoxMsgHist(req *gen_grpc.ChatGetBoxMsgHistReq) (*gen_grpc.ChatGetBoxMsgHistRes, error) {
	var err error
	var res gen_grpc.ChatGetBoxMsgHistRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("ChatGetBoxMsgHist: GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 检查用户是否在对话中
	switch x := req.GetPeerId().GetPeerIdUnion().(type) {
	case *gen_grpc.ChatPeerId_PeerUid:
		// 不作处理
	case *gen_grpc.ChatPeerId_GroupConvId:
		inConv, err := p.chat.IsUserInConv(x.GroupConvId, sessCtx.Uid)
		if err != nil {
			Log.Error("ChatGetBoxMsgHist: IsUserInConv: %s", err.Error())
			res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
			return &res, nil
		}
		if !inConv {
			res.ErrCode = gen_grpc.ErrCode_emErrCode_UserNotInConv
			return &res, nil
		}
	default:
		Log.Error("ChatGetBoxMsgHist: Unknown PeerId type")
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 获取聊天历史
	var peerId types.PeerId
	switch x := req.GetPeerId().GetPeerIdUnion().(type) {
	case *gen_grpc.ChatPeerId_PeerUid:
		peerId.PeerIdType = types.EmPeerIdType_PeerUid
		peerId.PeerUid = x.PeerUid
	case *gen_grpc.ChatPeerId_GroupConvId:
		peerId.PeerIdType = types.EmPeerIdType_GroupConvId
		peerId.GroupConvId = x.GroupConvId
	default:
		Log.Error("ChatGetBoxMsgHist: Unknown PeerId type")
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	var msgList []types.ChatMsg
	msgList, err = p.chat.GetConvMsgHist(peerId)
	if err != nil {
		Log.Error("ChatGetBoxMsgHist: GetConvMsgHist: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}
	for _, aMsg := range msgList {
		var aMsgApi gen_grpc.ChatMsg
		aMsgApi.MsgType = gen_grpc.ChatMsgType(aMsg.MsgType)
		aMsgApi.SentTsMs = aMsg.SentTsMs
		aMsgApi.SenderUid = aMsg.SenderUid
		aMsgApi.MsgContent = aMsg.MsgContent
		res.MsgList = append(res.MsgList, &aMsgApi)
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) ChatSendMsg(req *gen_grpc.ChatSendMsgReq) (*gen_grpc.ChatSendMsgRes, error) {
	var err error
	var res gen_grpc.ChatSendMsgRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("ChatSendMsg: GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 发送聊天消息
	var peerId types.PeerId
	switch x := req.GetBoxMsg().GetPeerId().GetPeerIdUnion().(type) {
	case *gen_grpc.ChatPeerId_PeerUid:
		peerId.PeerIdType = types.EmPeerIdType_PeerUid
		peerId.PeerUid = x.PeerUid
	case *gen_grpc.ChatPeerId_GroupConvId:
		peerId.PeerIdType = types.EmPeerIdType_GroupConvId
		peerId.GroupConvId = x.GroupConvId
	default:
		Log.Error("ChatSendMsg: Unknown PeerId type")
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	var msg types.ChatMsg
	msg.MsgType = types.EmChatMsgType(req.GetBoxMsg().GetMsg().GetMsgType())
	msg.SentTsMs = uint64(time.Now().UnixNano() / 1000000)
	msg.SenderUid = sessCtx.Uid
	msg.MsgContent = req.GetBoxMsg().GetMsg().GetMsgContent()

	err = p.chat.SendMsg(peerId, msg)
	if err != nil {
		Log.Error("ChatSendMsg: SendMsg: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) ChatCreateGroupConv(req *gen_grpc.ChatCreateGroupConvReq) (*gen_grpc.ChatCreateGroupConvRes, error) {
	var err error
	var res	gen_grpc.ChatCreateGroupConvRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("ChatCreateGroupConv: GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 创建群聊
	var convId uint64
	convId, err = p.chat.CreateGroupConv(sessCtx.Uid)
	if err != nil {
		Log.Error("ChatCreateGroupConv: CreateGroupConv: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ConvId = convId

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) ChatGetGroupConvList(req *gen_grpc.ChatGetGroupConvListReq) (*gen_grpc.ChatGetGroupConvListRes, error) {
	var err error
	var res	gen_grpc.ChatGetGroupConvListRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("ChatGetGroupConvList: GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 获取群聊列表
	var ConvIdList []uint64
	ConvIdList, err = p.chat.GetGroupConvList(sessCtx.Uid)
	if err != nil {
		Log.Error("ChatGetGroupConvList: GetGroupConvList: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ConvIdList = ConvIdList

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}