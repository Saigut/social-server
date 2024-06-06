package controller

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	uaParam.Username = req.GetUserid()
	uaParam.Passphase = req.GetPassphase()
	err = p.userMgmt.UserAuthenticate(&uaParam)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	// 检查已登录会话
	var sessCtx *SessCtxT
	sessCtx, err = p.sessMgmt.GetSessCtxByUsername(req.GetUserid())
	if err == nil {
		res.SessToken = string(sessCtx.SessId)
		res.Ret = 0
		return &res, nil
	}

	// 若未登录，创建新会话
	var sessId *SessIdT
	sessId, err = p.sessMgmt.CreateSess(req.GetUserid(), 2*3600)
	res.SessToken = string(*sessId)
	res.Ret = 0
	return &res, nil
}

func (p *ControllerT) SessUserLogout(sessId SessIdT, req *gen_grpc.SessUserLogoutReq) (*gen_grpc.SessUserLogoutRes, error) {
	var err error
	var res gen_grpc.SessUserLogoutRes

	// 获取会话
	_, err = p.sessMgmt.GetSessCtx(sessId)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	// 销毁会话
	err = p.sessMgmt.DeleteSess(sessId)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	res.Ret = 0
	return &res, nil
}

func (p *ControllerT) UmRegister(req *gen_grpc.UmRegisterReq) (*gen_grpc.UmRegisterRes, error) {
	var err error
	var res gen_grpc.UmRegisterRes

	// 用户名、密码、邮箱的合规性检查
	var validParam UmUserInfoValidateParam
	validParam.Username = req.GetUserid()
	validParam.Passphase = req.GetPassphase()
	validParam.Email = req.GetEmail()
	err = p.userMgmt.UserInfoValidate(&validParam)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	// 检查用户名是否已存在
	var existParam UmIsUsernameExistedParam
	existParam.Username = req.GetUserid()
	err = p.userMgmt.IsUsernameExisted(&existParam)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	// 创建用户
	var regParam UmRegisterParam
	regParam.Username = req.GetUserid()
	regParam.Passwd = req.GetPassphase()
	regParam.Email = req.GetEmail()
	err = p.userMgmt.Register(&regParam)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	res.Ret = 0
	return &res, nil
}

func (p *ControllerT) UmUnregister(sessId SessIdT, req *gen_grpc.UmUnregisterReq) (*gen_grpc.UmUnregisterRes, error) {
	var err error
	var res gen_grpc.UmUnregisterRes

	// 获取会话
	var sessCtx *SessCtxT
	sessCtx, err = p.sessMgmt.GetSessCtx(sessId)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	// 注销用户
	var unRegParam UmUnregisterParam
	unRegParam.Uid = sessCtx.Uid
	err = p.userMgmt.Unregister(&unRegParam)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	res.Ret = 0
	return &res, nil
}

func (p *ControllerT) UmAddFriends(sessId SessIdT, req *gen_grpc.UmAddFriendsReq) (*gen_grpc.UmAddFriendsRes, error) {
	var err error
	var res gen_grpc.UmAddFriendsRes

	// 获取会话
	var sessCtx *SessCtxT
	sessCtx, err = p.sessMgmt.GetSessCtx(sessId)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	// 添加好友
	var addFriendParam UmAddFriendsParam
	addFriendParam.Uid = sessCtx.Uid
	addFriendParam.FriendsUid = req.GetFriendsUid()
	err = p.userMgmt.AddFriends(&addFriendParam)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	res.Ret = 0
	return &res, nil
}

func (p *ControllerT) UmDelFriends(sessId SessIdT, req *gen_grpc.UmDelFriendsReq) (*gen_grpc.UmDelFriendsRes, error) {
	var err error
	var res gen_grpc.UmDelFriendsRes

	// 获取会话
	var sessCtx *SessCtxT
	sessCtx, err = p.sessMgmt.GetSessCtx(sessId)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	// 删除好友
	var delFriendParam UmDelFriendsParam
	delFriendParam.Uid = sessCtx.Uid
	delFriendParam.FriendsUid = req.GetFriendsUid()
	err = p.userMgmt.DelFriends(&delFriendParam)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	res.Ret = 0
	return &res, nil
}

func (p *ControllerT) UmListFriends(sessId SessIdT, req *gen_grpc.UmListFriendsReq) (*gen_grpc.UmListFriendsRes, error) {
	var err error
	var res gen_grpc.UmListFriendsRes

	// 获取会话
	var sessCtx *SessCtxT
	sessCtx, err = p.sessMgmt.GetSessCtx(sessId)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	// 获取好友列表
	var friendsUid []string
	var listFriendParam UmListFriendsParam
	listFriendParam.Uid = sessCtx.Uid
	friendsUid, err = p.userMgmt.ListFriends(&listFriendParam)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	for _, uid := range friendsUid {
		var friendInfo gen_grpc.UmFriendInfo
		friendInfo.Userid = uid
		friendInfo.Nickname = ""
		friendInfo.RemarkName = ""
		res.Friends = append(res.Friends, &friendInfo)
	}

	res.Ret = 0
	return &res, nil
}

func (p *ControllerT) ChatGetChatMsg(sessId SessIdT, req *gen_grpc.ChatGetChatMsgReq) (*gen_grpc.ChatGetChatMsgRes, error) {
	var err error
	var res gen_grpc.ChatGetChatMsgRes

	// 获取会话
	var sessCtx *SessCtxT
	sessCtx, err = p.sessMgmt.GetSessCtx(sessId)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	// 获取聊天消息
	var aMsg *ChatMsgOfConv
	aMsg, err = p.chat.GetChatMsg(sessCtx.Uid)

	var aMsgApi gen_grpc.ChatMsg
	aMsgApi.MsgContent = aMsg.Msg.MsgContent
	aMsgApi.MsgType = aMsg.Msg.MsgType
	aMsgApi.SenderUid = aMsg.Msg.SenderUid
	aMsgApi.SentTsMs = aMsg.Msg.SentTsMs
	res.Msgs = append(res.Msgs, &aMsgApi)

	res.Ret = 0
	return &res, nil
}

func (p *ControllerT) ChatGetChatMsgHistWith(sessId SessIdT, req *gen_grpc.ChatGetChatMsgHistWithReq) (*gen_grpc.ChatGetChatMsgHistWithRes, error) {
	var err error
	var res gen_grpc.ChatGetChatMsgHistWithRes

	// 获取会话
	//var sessCtx *SessCtxT
	_, err = p.sessMgmt.GetSessCtx(sessId)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	// 检查 req.GetConvId() 对话中是否包含 sessCtx.Uid

	// 获取聊天历史
	var aMsg *ChatMsg
	aMsg, err = p.chat.GetChatMsgHistWith(req.GetConvId())
	var aMsgApi gen_grpc.ChatMsg
	aMsgApi.MsgContent = aMsg.MsgContent
	aMsgApi.MsgType = aMsg.MsgType
	aMsgApi.SenderUid = aMsg.SenderUid
	aMsgApi.SentTsMs = aMsg.SentTsMs
	res.Msgs = append(res.Msgs, &aMsgApi)

	res.Ret = 0
	return &res, nil
}

func (p *ControllerT) ChatSendChatMsgTo(sessId SessIdT, req *gen_grpc.ChatSendChatMsgToReq) (*gen_grpc.ChatSendChatMsgToRes, error) {
	var err error
	var res gen_grpc.ChatSendChatMsgToRes

	// 获取会话
	var sessCtx *SessCtxT
	sessCtx, err = p.sessMgmt.GetSessCtx(sessId)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	// 发送聊天消息
	var msg ChatMsg
	msg.MsgContent = req.GetMsg().GetMsgContent()
	msg.MsgType = req.GetMsg().GetMsgType()
	msg.SenderUid = sessCtx.Uid
	msg.SentTsMs = uint64(time.Now().UnixNano() / 1000000)
	err = p.chat.SendChatMsgTo(&msg)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	res.Ret = 0
	return &res, nil
}

func (p *ControllerT) ChatGetChatConvId(sessId SessIdT, req *gen_grpc.ChatGetChatConvIdReq) (*gen_grpc.ChatGetChatConvIdRes, error) {
	var err error
	var res	gen_grpc.ChatGetChatConvIdRes

	// 获取会话
	//var sessCtx *SessCtxT
	_, err = p.sessMgmt.GetSessCtx(sessId)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	// 获取聊天的对话ID
	var convId uint64
	var param GetChatConvIdParam
	param.Uid1 = req.GetUid1()
	param.Uid2 = req.GetUid2()
	convId, err = p.chat.GetChatConvId(&param)
	if err != nil {
		res.Ret = -1
		return &res, nil
	}

	var convInfo gen_grpc.ChatConvInfo
	convInfo.ConvId = convId
	res.ConvInfo = &convInfo

	res.Ret = 0
	return &res, nil
}
