package controller

import (
	. "social_server/src/app/service/chat"
	. "social_server/src/app/service/sess_mgmt"
	. "social_server/src/app/service/user_mgmt"
	"social_server/src/gen/grpc"
	"time"
)


type ControllerT struct {
	userMgmt *UserMgmtT
	sessMgmt *SessMgmtT
	chat *ChatT
}

func (p *ControllerT) SessUserLogin(req *gen_grpc.SessUserLoginReq) (*gen_grpc.SessUserLoginRes, error) {
	var err error
	var res gen_grpc.SessUserLoginRes

	// 校验用户
	var uaParam UmUserAuthenticateParam
	uaParam.Username = req.GetUserName()
	uaParam.Passphase = req.GetPassphase()
	err = p.userMgmt.UserAuthenticate(&uaParam)
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 检查已登录会话
	var sessCtx *SessCtxT
	sessCtx, err = p.sessMgmt.GetSessCtxByUsername(req.GetUserName())
	if err == nil {
		res.SessId = string(sessCtx.SessId)
		res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
		return &res, nil
	}

	// 若未登录，创建新会话
	var sessId *SessIdT
	sessId, err = p.sessMgmt.CreateSess(req.GetUserName(), 2*3600)
	res.SessId = string(*sessId)
	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *ControllerT) SessUserLogout(req *gen_grpc.SessUserLogoutReq) (*gen_grpc.SessUserLogoutRes, error) {
	var err error
	var res gen_grpc.SessUserLogoutRes

	// 获取会话
	_, err = p.sessMgmt.GetSessCtx(SessIdT(req.GetSessId()))
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 销毁会话
	err = p.sessMgmt.DeleteSess(SessIdT(req.GetSessId()))
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *ControllerT) UmRegister(req *gen_grpc.UmRegisterReq) (*gen_grpc.UmRegisterRes, error) {
	var err error
	var res gen_grpc.UmRegisterRes

	// 用户名、密码、邮箱的合规性检查
	var validParam UmUserInfoValidateParam
	validParam.Username = req.GetUserName()
	validParam.Passphase = req.GetPassphase()
	validParam.Email = req.GetEmail()
	err = p.userMgmt.UserInfoValidate(&validParam)
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 检查用户名是否已存在
	err = p.userMgmt.IsUsernameExisted(req.GetUserName())
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 创建用户
	var regParam UmRegisterParam
	regParam.Username = req.GetUserName()
	regParam.Passwd = req.GetPassphase()
	regParam.Email = req.GetEmail()
	err = p.userMgmt.Register(&regParam)
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *ControllerT) UmUnregister(req *gen_grpc.UmUnregisterReq) (*gen_grpc.UmUnregisterRes, error) {
	var err error
	var res gen_grpc.UmUnregisterRes

	// 获取会话
	var sessCtx *SessCtxT
	sessCtx, err = p.sessMgmt.GetSessCtx(SessIdT(req.GetSessId()))
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 注销用户
	var unRegParam UmUnregisterParam
	unRegParam.Uid = sessCtx.Uid
	err = p.userMgmt.Unregister(&unRegParam)
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *ControllerT) UmAddFriend(req *gen_grpc.UmAddFriendReq) (*gen_grpc.UmAddFriendRes, error) {
	var err error
	var res gen_grpc.UmAddFriendRes

	// 获取会话
	var sessCtx *SessCtxT
	sessCtx, err = p.sessMgmt.GetSessCtx(SessIdT(req.GetSessId()))
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 添加好友
	var addFriendParam UmAddFriendsParam
	addFriendParam.Uid = sessCtx.Uid
	addFriendParam.FriendUid = req.GetFriend().GetUid()
	err = p.userMgmt.AddFriends(&addFriendParam)
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *ControllerT) UmDelFriends(req *gen_grpc.UmDelFriendReq) (*gen_grpc.UmDelFriendRes, error) {
	var err error
	var res gen_grpc.UmDelFriendRes

	// 获取会话
	var sessCtx *SessCtxT
	sessCtx, err = p.sessMgmt.GetSessCtx(SessIdT(req.GetSessId()))
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 删除好友
	var delFriendParam UmDelFriendsParam
	delFriendParam.Uid = sessCtx.Uid
	delFriendParam.FriendUid = req.GetFriendUid()
	err = p.userMgmt.DelFriends(&delFriendParam)
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *ControllerT) UmListFriends(req *gen_grpc.UmGetFriendListReq) (*gen_grpc.UmGetFriendListRes, error) {
	var err error
	var res gen_grpc.UmGetFriendListRes

	// 获取会话
	var sessCtx *SessCtxT
	sessCtx, err = p.sessMgmt.GetSessCtx(SessIdT(req.GetSessId()))
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 获取好友列表
	var friendUidList []uint64
	friendUidList, err = p.userMgmt.GetFriendList(sessCtx.Uid)
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	for _, uid := range friendUidList {
		var friendInfo gen_grpc.UmFriendInfo
		friendInfo.Uid = uid
		friendInfo.NickName = ""
		friendInfo.RemarkName = ""
		res.FriendList = append(res.FriendList, &friendInfo)
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *ControllerT) ChatGetMsgList(req *gen_grpc.ChatGetMsgListReq) (*gen_grpc.ChatGetMsgListRes, error) {
	var err error
	var res gen_grpc.ChatGetMsgListRes

	// 获取会话
	var sessCtx *SessCtxT
	sessCtx, err = p.sessMgmt.GetSessCtx(SessIdT(req.GetSessId()))
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 获取聊天消息
	var msgList []ChatMsgOfConv
	msgList, err = p.chat.GetChatMsgList(sessCtx.Uid)
	for _, aConvMsg := range msgList {
		var aBoxMsgApi gen_grpc.ChatBoxMsg
		if aConvMsg.PeerId.PeerIdType == 0 {
			aBoxMsgApi.PeerId.PeerIdUnion = &gen_grpc.ChatPeerId_PeerUid{PeerUid: aConvMsg.PeerId.PeerUid}
		} else {
			aBoxMsgApi.PeerId.PeerIdUnion = &gen_grpc.ChatPeerId_GroupConvId{GroupConvId: aConvMsg.PeerId.GroupConvId}
		}
		aBoxMsgApi.Msg.MsgType = gen_grpc.ChatMsgType(aConvMsg.Msg.MsgType)
		aBoxMsgApi.Msg.SentTsMs = aConvMsg.Msg.SentTsMs
		aBoxMsgApi.Msg.SenderUid = aConvMsg.Msg.SenderUid
		aBoxMsgApi.Msg.MsgContent = aConvMsg.Msg.MsgContent
		res.MsgList = append(res.MsgList, &aBoxMsgApi)
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *ControllerT) ChatGetBoxMsgHist(req *gen_grpc.ChatGetBoxMsgHistReq) (*gen_grpc.ChatGetBoxMsgHistRes, error) {
	var err error
	var res gen_grpc.ChatGetBoxMsgHistRes

	// 获取会话
	var sessCtx *SessCtxT
	sessCtx, err = p.sessMgmt.GetSessCtx(SessIdT(req.GetSessId()))
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 检查用户是否在对话中
	switch x := req.GetPeerId().GetPeerIdUnion().(type) {
	case *gen_grpc.ChatPeerId_PeerUid:
		// 不作处理
	case *gen_grpc.ChatPeerId_GroupConvId:
		if !p.chat.IsUserInConv(x.GroupConvId, sessCtx.Uid) {
			res.ErrCode = gen_grpc.ErrCode_emErrCode_UserNotInConv
			return &res, nil
		}
	default:
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 获取聊天历史
	var peerId PeerId
	switch x := req.GetPeerId().GetPeerIdUnion().(type) {
	case *gen_grpc.ChatPeerId_PeerUid:
		peerId.PeerIdType = 0
		peerId.PeerUid = x.PeerUid
	case *gen_grpc.ChatPeerId_GroupConvId:
		peerId.PeerIdType = 1
		peerId.GroupConvId = x.GroupConvId
	default:
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	var msgList []ChatMsg
	msgList, err = p.chat.GetConvMsgHist(peerId)
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

func (p *ControllerT) ChatSendMsg(req *gen_grpc.ChatSendMsgReq) (*gen_grpc.ChatSendMsgRes, error) {
	var err error
	var res gen_grpc.ChatSendMsgRes

	// 获取会话
	var sessCtx *SessCtxT
	sessCtx, err = p.sessMgmt.GetSessCtx(SessIdT(req.GetSessId()))
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 发送聊天消息
	var peerId PeerId
	switch x := req.GetBoxMsg().GetPeerId().GetPeerIdUnion().(type) {
	case *gen_grpc.ChatPeerId_PeerUid:
		peerId.PeerIdType = 0
		peerId.PeerUid = x.PeerUid
	case *gen_grpc.ChatPeerId_GroupConvId:
		peerId.PeerIdType = 1
		peerId.GroupConvId = x.GroupConvId
	default:
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	var msg ChatMsg
	msg.MsgType = EmChatMsgType(req.GetBoxMsg().GetMsg().GetMsgType())
	msg.SentTsMs = uint64(time.Now().UnixNano() / 1000000)
	msg.SenderUid = sessCtx.Uid
	msg.MsgContent = req.GetBoxMsg().GetMsg().GetMsgContent()

	err = p.chat.SendChatMsgTo(peerId, msg)
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *ControllerT) ChatCreateGroupConv(req *gen_grpc.ChatCreateGroupConvReq) (*gen_grpc.ChatCreateGroupConvRes, error) {
	var err error
	var res	gen_grpc.ChatCreateGroupConvRes

	// 获取会话
	var sessCtx *SessCtxT
	sessCtx, err = p.sessMgmt.GetSessCtx(SessIdT(req.GetSessId()))
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 创建群聊
	var convId uint64
	convId, err = p.chat.CreateGroupConv(sessCtx.Uid)
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ConvId = convId

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *ControllerT) ChatGetGroupConvList(req *gen_grpc.ChatGetGroupConvListReq) (*gen_grpc.ChatGetGroupConvListRes, error) {
	var err error
	var res	gen_grpc.ChatGetGroupConvListRes

	// 获取会话
	var sessCtx *SessCtxT
	sessCtx, err = p.sessMgmt.GetSessCtx(SessIdT(req.GetSessId()))
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 获取群聊列表
	var ConvIdList []uint64
	ConvIdList, err = p.chat.GetGroupConvList(sessCtx.Uid)
	if err != nil {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ConvIdList = ConvIdList

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}