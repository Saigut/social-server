package core

import (
	"errors"
	"github.com/joho/godotenv"
	"log"
	"os"
	"social_server/src/app/common/proj_err"
	"social_server/src/app/common/types"
	"social_server/src/app/common/utils"
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
	sessTimoutS uint64
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
		sessTimoutS: 60 * 60 * 2, // 2小时
	}

	return p
}

func (p *Core) SessUserLogin(req *gen_grpc.SessUserLoginReq) (*gen_grpc.SessUserLoginRes, error) {
	var err error
	var res gen_grpc.SessUserLoginRes

	// 校验用户
	var uaParam types.UmUserAuthenticateParam
	uaParam.Username = req.GetUsername()
	uaParam.Passphase = utils.CalPassHash(req.GetPassword())
	var pass bool
	pass, err = p.userMgmt.UserAuthenticate(&uaParam)
	if err != nil {
		Log.Error("UserAuthenticate: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}
	if !pass {
		Log.Error("User auth failed")
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UserFailedToAuth
		return &res, nil
	}

	// 获取用户信息
	var userInfo *types.UmUserInfo
	userInfo, err = p.userMgmt.UserGetInfoByUsername(req.GetUsername())
	if err != nil {
		Log.Error("UserGetInfoByUsername: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 创建新会话
	var sessId types.SessId
	sessId, err = p.sessMgmt.CreateSess(req.GetUsername(), userInfo.Uid, p.sessTimoutS)
	if err != nil {
		Log.Error("CreateSess: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}
	Log.Debug("SessId: %s", res.SessId)

	res.SessId = string(sessId)
	res.Uid = userInfo.Uid
	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok

	return &res, nil
}

func (p *Core) SessUserLogout(req *gen_grpc.SessUserLogoutReq) (*gen_grpc.SessUserLogoutRes, error) {
	var err error
	var res gen_grpc.SessUserLogoutRes

	// 获取会话
	_, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 销毁会话
	err = p.sessMgmt.DeleteSess(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("DeleteSess: %s", err.Error())
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
		Log.Error("UserInfoValidate: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 检查用户名是否已存在
	var isExist bool
	isExist, err = p.userMgmt.IsUsernameExisted(req.GetUsername())
	if err != nil {
		Log.Error("UserIsUsernameExisted: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}
	if isExist {
		Log.Warn("User already registered")
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UserAlreadyRegistered
		return &res, nil
	}

	// 创建用户
	var regParam types.UmRegisterParam
	regParam.Username = req.GetUsername()
	regParam.Passwd = utils.CalPassHash(req.GetPassword())
	regParam.Nickname = req.GetNickname()
	regParam.Email = req.GetEmail()
	regParam.Avatar = req.GetAvatar()
	err = p.userMgmt.Register(&regParam)
	if err != nil {
		Log.Error("Register: %s", err.Error())
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
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 注销用户
	var unRegParam types.UmUnregisterParam
	unRegParam.Uid = sessCtx.Uid
	err = p.userMgmt.Unregister(&unRegParam)
	if err != nil {
		Log.Error("Unregister: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 销毁用户会话
	err = p.sessMgmt.DeleteUserSess(sessCtx.Uid)
	if err != nil {
		Log.Error("DeleteUserSess: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmUserUpdateInfo(req *gen_grpc.UmUserUpdateInfoReq) (*gen_grpc.UmUserUpdateInfoRes, error) {
	var err error
	var res gen_grpc.UmUserUpdateInfoRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	var password = ""
	if req.GetPassword() != "" {
		password = utils.CalPassHash(req.GetPassword())
	}
	var newPassword = ""
	if req.GetNewPassword() != "" {
		if !p.userMgmt.ValidatePassword(req.GetNewPassword()) {
			Log.Error("Password validate failed: %s")
			res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
			return &res, nil
		}
		newPassword = utils.CalPassHash(req.GetNewPassword())
	}
	// 更新用户信息
	err = p.userMgmt.UserUpdateInfo(sessCtx.Uid, req.GetNickname(),
		req.GetEmail(), req.GetAvatar(), password, newPassword)
	if err != nil {
		Log.Error("UserUpdateInfo: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmContactGetList(req *gen_grpc.UmContactGetListReq) (*gen_grpc.UmContactGetListRes, error) {
	var err error
	var res gen_grpc.UmContactGetListRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 获取好友列表
	var contactUidList []uint64
	contactUidList, err = p.userMgmt.ContactGetList(sessCtx.Uid)
	if err != nil {
		Log.Error("ContactGetList: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	for _, uid := range contactUidList {
		// 获取联系人信息
		var userInfo *types.UmUserInfo
		userInfo, err = p.userMgmt.ContactGetInfo(uid)
		if err != nil {
			Log.Error("ContactGetInfo: %s", err.Error())
			res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
			return &res, nil
		}

		isMutualContact, remarkName, err := p.userMgmt.ContactGetRelation(sessCtx.Uid, uid)
		if err != nil {
			Log.Error("ContactGetRelation: %s", err.Error())
			res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
			return &res, nil
		}

		var contactInfo gen_grpc.UmContactInfo
		contactInfo.Uid = userInfo.Uid
		contactInfo.Username = userInfo.Username
		contactInfo.Nickname = userInfo.Nickname
		contactInfo.Avatar = userInfo.Avatar
		contactInfo.Email = userInfo.Email
		contactInfo.NoteName = remarkName
		contactInfo.IsMutualContact = isMutualContact
		res.ContactList = append(res.ContactList, &contactInfo)
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmContactGetInfo(req *gen_grpc.UmContactGetInfoReq) (*gen_grpc.UmContactGetInfoRes, error) {
	var err error
	var res gen_grpc.UmContactGetInfoRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 查找用户
	var userInfo *types.UmUserInfo
	userInfo, err = p.userMgmt.ContactGetInfo(req.GetUserId())
	if err != nil {
		Log.Error("ContactFind: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	isMutualContact, remarkName, err := p.userMgmt.ContactGetRelation(sessCtx.Uid, req.GetUserId())
	if err != nil {
		Log.Error("ContactGetRelation: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	var contactInfo gen_grpc.UmContactInfo
	contactInfo.Uid = userInfo.Uid
	contactInfo.Username = userInfo.Username
	contactInfo.Nickname = userInfo.Nickname
	contactInfo.NoteName = remarkName
	contactInfo.Email = userInfo.Email
	contactInfo.Avatar = userInfo.Avatar
	contactInfo.IsMutualContact = isMutualContact

	res.UserInfo = &contactInfo

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmContactFind(req *gen_grpc.UmContactFindReq) (*gen_grpc.UmContactFindRes, error) {
	var err error
	var res gen_grpc.UmContactFindRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 查找用户
	var userInfo *types.UmUserInfo
	userInfo, err = p.userMgmt.ContactFind(req.GetUsername())
	if err != nil {
		Log.Error("ContactFind: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	isMutualContact, remarkName, err := p.userMgmt.ContactGetRelation(sessCtx.Uid, userInfo.Uid)
	if err != nil {
		Log.Error("ContactGetRelation: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	var contactInfo gen_grpc.UmContactInfo
	contactInfo.Uid = userInfo.Uid
	contactInfo.Username = userInfo.Username
	contactInfo.Nickname = userInfo.Nickname
	contactInfo.NoteName = remarkName
	contactInfo.Email = userInfo.Email
	contactInfo.Avatar = userInfo.Avatar
	contactInfo.IsMutualContact = isMutualContact

	res.UserInfo = &contactInfo

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmContactAddRequest(req *gen_grpc.UmContactAddRequestReq) (*gen_grpc.UmContactAddRequestRes, error) {
	var err error
	var res gen_grpc.UmContactAddRequestRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 发送好友申请消息
	var msg types.ChatMsgOfConv
	msg.ReceiverId.PeerIdType = types.EmPeerIdType_Uid
	msg.ReceiverId.Uid = req.GetContactUid()
	msg.Msg.SenderUid = sessCtx.Uid
	msg.Msg.MsgType = gen_grpc.ChatMsgType_emChatMsgType_ContactAddReq
	err = p.chat.SendMsg(msg)
	if err != nil {
		Log.Error("SendMsg: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmContactAccept(req *gen_grpc.UmContactAcceptReq) (*gen_grpc.UmContactAcceptRes, error) {
	var err error
	var res gen_grpc.UmContactAcceptRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 同意好友请求
	err = p.userMgmt.ContactAccept(sessCtx.Uid, req.GetContactUid())
	if err != nil {
		Log.Error("ContactAccept: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 发送通过好友消息
	var msg types.ChatMsgOfConv
	msg.ReceiverId.PeerIdType = types.EmPeerIdType_Uid
	msg.ReceiverId.Uid = req.GetContactUid()
	msg.Msg.SenderUid = sessCtx.Uid
	msg.Msg.MsgType = gen_grpc.ChatMsgType_emChatMsgType_ContactAdded
	msg.Msg.MsgContent = "我通过了你的朋友验证请求，现在我们可以开始聊天了"
	err = p.chat.SendMsg(msg)
	if err != nil {
		Log.Error("SendMsg: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmContactReject(req *gen_grpc.UmContactRejectReq) (*gen_grpc.UmContactRejectRes, error) {
	var err error
	var res gen_grpc.UmContactRejectRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 拒绝好友请求
	err = p.userMgmt.ContactReject(sessCtx.Uid, req.GetContactUid())
	if err != nil {
		Log.Error("ContactReject: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 发送拒绝好友消息
	var msg types.ChatMsgOfConv
	msg.ReceiverId.PeerIdType = types.EmPeerIdType_Uid
	msg.ReceiverId.Uid = req.GetContactUid()
	msg.Msg.SenderUid = sessCtx.Uid
	msg.Msg.MsgType = gen_grpc.ChatMsgType_emChatMsgType_ContactRejected
	err = p.chat.SendMsg(msg)
	if err != nil {
		Log.Error("SendMsg: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmContactDel(req *gen_grpc.UmContactDelReq) (*gen_grpc.UmContactDelRes, error) {
	var err error
	var res gen_grpc.UmContactDelRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 删除好友
	err = p.userMgmt.ContactDel(sessCtx.Uid, req.GetContactUid())
	if err != nil {
		Log.Error("ContactDel: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 发送删除好友消息
	var msg types.ChatMsgOfConv
	msg.ReceiverId.PeerIdType = types.EmPeerIdType_Uid
	msg.ReceiverId.Uid = req.GetContactUid()
	msg.Msg.SenderUid = sessCtx.Uid
	msg.Msg.MsgType = gen_grpc.ChatMsgType_emChatMsgType_ContactDeleted
	err = p.chat.SendMsg(msg)
	if err != nil {
		Log.Error("SendMsg: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmGroupGetList(req *gen_grpc.UmGroupGetListReq) (*gen_grpc.UmGroupGetListRes, error) {
	var err error
	var res	gen_grpc.UmGroupGetListRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 获取群聊列表
	var groupIdList []uint64
	groupIdList, err = p.userMgmt.GroupGetList(sessCtx.Uid)
	if err != nil {
		Log.Error("GroupGetList: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	for _, ConvId := range groupIdList {
		// 获取群聊信息
		var groupInfo *types.UmGroupInfo
		groupInfo, err = p.userMgmt.GroupGetInfo(ConvId)
		if err != nil {
			Log.Error("GroupGetInfo: %s, ConvId: %v", err.Error(), ConvId)
			res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
			return &res, nil
		}

		var grpcGroupInfo gen_grpc.UmGroupInfo
		grpcGroupInfo.GroupId = groupInfo.GroupId
		grpcGroupInfo.GroupName = groupInfo.GroupName
		grpcGroupInfo.OwnerUid = groupInfo.OwnerUid
		grpcGroupInfo.CreateTsMs = groupInfo.CreateTsMs

		res.GroupList = append(res.GroupList, &grpcGroupInfo)
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmGroupGetInfo(req *gen_grpc.UmGroupGetInfoReq) (*gen_grpc.UmGroupGetInfoRes, error) {
	var err error
	var res gen_grpc.UmGroupGetInfoRes

	// 获取会话
	_, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 获取群组信息
	var groupInfo *types.UmGroupInfo
	groupInfo, err = p.userMgmt.GroupGetInfo(req.GetGroupId())
	if err != nil {
		Log.Error("GroupGetInfo: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.GroupInfo = &gen_grpc.UmGroupInfo{
		GroupId: groupInfo.GroupId,
		GroupName: groupInfo.GroupName,
		OwnerUid: groupInfo.OwnerUid,
		Avatar: groupInfo.Avatar,
		MemCount: groupInfo.MemCount,
		CreateTsMs: groupInfo.CreateTsMs,
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

// UmGroupUpdateInfo
func (p *Core) UmGroupUpdateInfo(req *gen_grpc.UmGroupUpdateInfoReq) (*gen_grpc.UmGroupUpdateInfoRes, error) {
	var err error
	var res gen_grpc.UmGroupUpdateInfoRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 判断是否是群主
	isOwner, err := p.userMgmt.GroupIsOwner(req.GetGroupId(), sessCtx.Uid)
	if err != nil {
		Log.Error("GroupIsOwner: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}
	if !isOwner {
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 更新群信息
	err = p.userMgmt.GroupUpdateInfo(req.GetGroupId(), req.GetGroupName(), req.GetAvatar())
	if err != nil {
		Log.Error("GroupUpdateInfo: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmGroupFind(req *gen_grpc.UmGroupFindReq) (*gen_grpc.UmGroupFindRes, error) {
	var err error
	var res gen_grpc.UmGroupFindRes

	// 获取会话
	_, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 查找群组
	var groupInfo *types.UmGroupInfo
	groupInfo, err = p.userMgmt.GroupFind(req.GetGroupId())
	if err != nil {
		Log.Error("GroupFind: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.GroupInfo = &gen_grpc.UmGroupInfo{
		GroupId: groupInfo.GroupId,
		GroupName: groupInfo.GroupName,
		OwnerUid: groupInfo.OwnerUid,
		CreateTsMs: groupInfo.CreateTsMs,
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmGroupCreate(req *gen_grpc.UmGroupCreateReq) (*gen_grpc.UmGroupCreateRes, error) {
	var err error
	var res	gen_grpc.UmGroupCreateRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 创建群聊
	var groupId uint64
	groupId, err = p.userMgmt.GroupCreate(sessCtx.Uid, req.GroupName)
	if err != nil {
		Log.Error("GroupCreate: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 发送创建群聊消息
	var msg types.ChatMsgOfConv
	msg.ReceiverId.PeerIdType = types.EmPeerIdType_GroupId
	msg.ReceiverId.GroupId = groupId
	msg.Msg.SenderUid = sessCtx.Uid
	msg.Msg.MsgType = gen_grpc.ChatMsgType_emChatMsgType_GroupCreated
	err = p.chat.SendMsg(msg)
	if err != nil {
		Log.Error("SendMsg: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.GroupId = groupId

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

// TODO：先标记、再通知、再删除
func (p *Core) UmGroupDelete(req *gen_grpc.UmGroupDeleteReq) (*gen_grpc.UmGroupDeleteRes, error) {
	var err error
	var res	gen_grpc.UmGroupDeleteRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 分配群聊消息序列号
	seqId, err := p.chat.AllocateGroupSeqId(req.GetGroupId())
	if err != nil {
		Log.Error("AllocateGroupSeqId: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 获取成员列表
	var memList []uint64
	memList, err = p.userMgmt.GroupGetMemList(req.GetGroupId())
	if err != nil {
		Log.Error("GroupGetMemList: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 解散群聊
	err = p.userMgmt.GroupDelete(sessCtx.Uid, req.GetGroupId())
	if err != nil {
		Log.Error("GroupDelete: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 遍历群成员发送解散消息
	for _, uid := range memList {
		var msg types.ChatMsgOfConv
		msg.ReceiverId.PeerIdType = types.EmPeerIdType_GroupId
		msg.ReceiverId.GroupId = req.GetGroupId()
		msg.Msg.SenderUid = sessCtx.Uid
		msg.Msg.MsgType = gen_grpc.ChatMsgType_emChatMsgType_GroupDeleted
		msg.ConvMsgId = seqId
		err = p.chat.SendMsgToUser(uid, msg)
		if err != nil {
			Log.Error("SendMsg: %s", err.Error())
		}
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmGroupGetMemList(req *gen_grpc.UmGroupGetMemListReq) (*gen_grpc.UmGroupGetMemListRes, error) {
	var err error
	var res gen_grpc.UmGroupGetMemListRes

	// 获取会话
	_, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 获取群成员列表
	var UidList []uint64
	UidList, err = p.userMgmt.GroupGetMemList(req.GetGroupId())
	if err != nil {
		Log.Error("GroupGetMemList: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}
	res.MemUidList = UidList

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmGroupJoinRequest(req *gen_grpc.UmGroupJoinRequestReq) (*gen_grpc.UmGroupJoinRequestRes, error) {
	var err error
	var res gen_grpc.UmGroupJoinRequestRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 向每个管理员发送入群申请消息
	var msg types.ChatMsgOfConv
	msg.ReceiverId.PeerIdType = types.EmPeerIdType_GroupId
	msg.ReceiverId.GroupId = req.GetGroupId()
	msg.Msg.SenderUid = sessCtx.Uid
	msg.Msg.MsgType = gen_grpc.ChatMsgType_emChatMsgType_GroupJoinReq
	err = p.chat.SendMsgToAdmins(msg)
	if err != nil {
		Log.Error("SendMsgToAdmins: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmGroupAccept(req *gen_grpc.UmGroupAcceptReq) (*gen_grpc.UmGroupAcceptRes, error) {
	var err error
	var res gen_grpc.UmGroupAcceptRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
	}

	// 允许加入
	err = p.userMgmt.GroupAccept(req.GetGroupId(), sessCtx.Uid, req.GetUid())
	if err != nil {
		Log.Error("GroupAccept: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 向群内发送入群消息
	var msg types.ChatMsgOfConv
	msg.ReceiverId.PeerIdType = types.EmPeerIdType_GroupId
	msg.ReceiverId.GroupId = req.GetGroupId()
	msg.Msg.SenderUid = req.GetUid()
	msg.Msg.MsgType = gen_grpc.ChatMsgType_emChatMsgType_GroupUserJoined
	err = p.chat.SendMsg(msg)
	if err != nil {
		Log.Error("SendMsg: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmGroupReject(req *gen_grpc.UmGroupRejectReq) (*gen_grpc.UmGroupRejectRes, error) {
	var err error
	var res gen_grpc.UmGroupRejectRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
	}

	// 拒绝加入请求
	err = p.userMgmt.GroupReject(req.GetGroupId(), sessCtx.Uid, req.GetUid())
	if err != nil {
		Log.Error("GroupReject: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 向用户发送拒绝消息
	var msg types.ChatMsgOfConv
	msg.ReceiverId.PeerIdType = types.EmPeerIdType_Uid
	msg.ReceiverId.Uid = req.GetUid()
	msg.Msg.SenderUid = sessCtx.Uid
	msg.Msg.MsgType = gen_grpc.ChatMsgType_emChatMsgType_GroupRejected
	err = p.chat.SendMsg(msg)
	if err != nil {
		Log.Error("SendMsg: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmGroupLeave(req *gen_grpc.UmGroupLeaveReq) (*gen_grpc.UmGroupLeaveRes, error) {
	var err error
	var res gen_grpc.UmGroupLeaveRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 分配群聊消息序列号
	seqId, err := p.chat.AllocateGroupSeqId(req.GetGroupId())
	if err != nil {
		Log.Error("AllocateGroupSeqId: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 获取成员列表
	var memList []uint64
	memList, err = p.userMgmt.GroupGetMemList(req.GetGroupId())
	if err != nil {
		Log.Error("GroupGetMemList: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 离开群组
	err = p.userMgmt.GroupLeave(req.GetGroupId(), sessCtx.Uid)
	if err != nil {
		Log.Error("GroupLeave: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 发送退出群聊消息
	for _, uid := range memList {
		var msg types.ChatMsgOfConv
		msg.ReceiverId.PeerIdType = types.EmPeerIdType_GroupId
		msg.ReceiverId.GroupId = req.GetGroupId()
		msg.Msg.SenderUid = sessCtx.Uid
		msg.Msg.MsgType = gen_grpc.ChatMsgType_emChatMsgType_GroupUserLeft
		msg.ConvMsgId = seqId
		err = p.chat.SendMsgToUser(uid, msg)
		if err != nil {
			Log.Error("SendMsg: %s", err.Error())
		}
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmGroupAddMem(req *gen_grpc.UmGroupAddMemReq) (*gen_grpc.UmGroupAddMemRes, error) {
	var err error
	var res gen_grpc.UmGroupAddMemRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 添加群成员
	err = p.userMgmt.GroupAddMem(req.GetGroupId(), sessCtx.Uid, req.GetUid())
	if err != nil {
		Log.Error("GroupAddMem: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 发送添加群成员消息
	var msg types.ChatMsgOfConv
	msg.ReceiverId.PeerIdType = types.EmPeerIdType_GroupId
	msg.ReceiverId.GroupId = req.GetGroupId()
	msg.Msg.SenderUid = req.GetUid()
	msg.Msg.MsgType = gen_grpc.ChatMsgType_emChatMsgType_GroupUserJoined
	err = p.chat.SendMsg(msg)
	if err != nil {
		Log.Error("SendMsg: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmGroupDelMem(req *gen_grpc.UmGroupDelMemReq) (*gen_grpc.UmGroupDelMemRes, error) {
	var err error
	var res gen_grpc.UmGroupDelMemRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 分配群聊消息序列号
	seqId, err := p.chat.AllocateGroupSeqId(req.GetGroupId())
	if err != nil {
		Log.Error("AllocateGroupSeqId: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 获取成员列表
	var memList []uint64
	memList, err = p.userMgmt.GroupGetMemList(req.GetGroupId())
	if err != nil {
		Log.Error("GroupGetMemList: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 移除群成员
	err = p.userMgmt.GroupDelMem(req.GetGroupId(), sessCtx.Uid, req.GetUid())
	if err != nil {
		Log.Error("GroupDelMem: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 发送删除群成员消息
	for _, uid := range memList {
		var msg types.ChatMsgOfConv
		msg.ReceiverId.PeerIdType = types.EmPeerIdType_GroupId
		msg.ReceiverId.GroupId = req.GetGroupId()
		msg.Msg.SenderUid = req.GetUid()
		msg.Msg.MsgType = gen_grpc.ChatMsgType_emChatMsgType_GroupUserRemoved
		msg.ConvMsgId = seqId
		err = p.chat.SendMsgToUser(uid, msg)
		if err != nil {
			Log.Error("SendMsg: %s", err.Error())
		}
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) UmGroupUpdateMem(req *gen_grpc.UmGroupUpdateMemReq) (*gen_grpc.UmGroupUpdateMemRes, error) {
	var err error
	var res gen_grpc.UmGroupUpdateMemRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 更新群成员
	err = p.userMgmt.GroupUpdateMem(req.GetGroupId(), sessCtx.Uid, req.GetUid(), uint(req.GetRole()))
	if err != nil {
		Log.Error("GroupUpdateMem: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
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
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	var convMsg types.ChatMsgOfConv
	switch x := req.GetConvMsg().GetReceiverId().GetPeerIdUnion().(type) {
	case *gen_grpc.ChatPeerId_Uid:
		// 判断是否为好友 ContactGetRelation
		isMutualContact, _, err := p.userMgmt.ContactGetRelation(sessCtx.Uid, x.Uid)
		if err != nil {
			Log.Error("ContactGetRelation: %s", err.Error())
			res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
			return &res, nil
		}
		if !isMutualContact {
			Log.Error("Not mutual contact")
			res.ErrCode = gen_grpc.ErrCode_emErrCode_IsNotContact
			return &res, nil
		}
		convMsg.ReceiverId.PeerIdType = types.EmPeerIdType_Uid
		convMsg.ReceiverId.Uid = x.Uid
	case *gen_grpc.ChatPeerId_GroupId:
		// 判断是否为群成员
		inGroup, err := p.userMgmt.GroupIsMem(x.GroupId, sessCtx.Uid)
		if err != nil {
			Log.Error("GroupIsMem: %s", err.Error())
			res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
			return &res, nil
		}
		if !inGroup {
			Log.Error("Not in group")
			res.ErrCode = gen_grpc.ErrCode_emErrCode_UserNotInGroup
			return &res, nil
		}
		convMsg.ReceiverId.PeerIdType = types.EmPeerIdType_GroupId
		convMsg.ReceiverId.GroupId = x.GroupId
	default:
		Log.Error("Unknown ReceiverId type")
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	convMsg.Msg.SentTsMs = uint64(time.Now().UnixNano() / 1000000)
	convMsg.Msg.SenderUid = sessCtx.Uid
	convMsg.Msg.MsgType = req.GetConvMsg().GetMsg().GetMsgType()
	convMsg.Msg.MsgContent = req.GetConvMsg().GetMsg().GetMsgContent()
	convMsg.Msg.ReadMsgId = req.GetConvMsg().GetMsg().GetReadMsgId()

	convMsg.RandMsgId = req.GetConvMsg().GetRandMsgId()

	err = p.chat.SendMsg(convMsg)
	if err != nil {
		Log.Error("ChatSendMsg: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) ChatMarkRead(req *gen_grpc.ChatMarkReadReq) (*gen_grpc.ChatMarkReadRes, error) {
	var err error
	var res gen_grpc.ChatMarkReadRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	var convMsg types.ChatMsgOfConv
	switch x := req.GetConvId().GetPeerIdUnion().(type) {
	case *gen_grpc.ChatPeerId_Uid:
		convMsg.ReceiverId.PeerIdType = types.EmPeerIdType_Uid
		convMsg.ReceiverId.Uid = x.Uid
	case *gen_grpc.ChatPeerId_GroupId:
		convMsg.ReceiverId.PeerIdType = types.EmPeerIdType_GroupId
		convMsg.ReceiverId.GroupId = x.GroupId
	default:
		Log.Error("Unknown ReceiverId type")
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	convMsg.Msg.SentTsMs = uint64(time.Now().UnixNano() / 1000000)
	convMsg.Msg.SenderUid = sessCtx.Uid
	convMsg.Msg.MsgType = gen_grpc.ChatMsgType_emChatMsgType_MarkRead
	convMsg.Msg.ReadMsgId = req.GetReadMsgId()

	err = p.chat.SendMsg(convMsg)
	if err != nil {
		Log.Error("ChatSendMsg: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}

func (p *Core) GetUpdateList(req *gen_grpc.GetUpdateListReq) (*gen_grpc.GetUpdateListRes, error) {
	var err error
	var res gen_grpc.GetUpdateListRes

	// 获取会话
	var sessCtx *types.SessCtx
	sessCtx, err = p.sessMgmt.GetSessCtx(types.SessId(req.GetSessId()))
	if err != nil {
		Log.Error("GetSessCtx: %s", err.Error())
		res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
		return &res, nil
	}

	// 会话续期
	err = p.sessMgmt.RenewSessCtx(sessCtx, p.sessTimoutS)
	if err != nil {
		Log.Warn("RenewSessCtx: %v", err)
	}

	// 获取聊天消息
	var msgList []types.ChatMsgOfConv
	msgList, err = p.chat.GetChatMsgList(sessCtx.Uid, req.GetLocalSeqId())
	if err != nil {
		Log.Error("ChatGetMsgList: %s", err.Error())
		if errors.Is(err, proj_err.ErrTimeout) {
			res.ErrCode = gen_grpc.ErrCode_emErrCode_Timeout
			return &res, nil
		} else {
			res.ErrCode = gen_grpc.ErrCode_emErrCode_UnknownErr
			return &res, nil
		}
	}
	for _, aConvMsg := range msgList {
		aBoxMsgApi := &gen_grpc.ChatConvMsg{
			SeqId: aConvMsg.SeqId,
			ConvMsgId:  aConvMsg.ConvMsgId,
			RandMsgId: aConvMsg.RandMsgId,
			ReceiverId: &gen_grpc.ChatPeerId{},
			Msg:    &gen_grpc.ChatMsg{},
			IsRead: aConvMsg.IsRead,
			Status: aConvMsg.Status,
		}
		if aConvMsg.ReceiverId.PeerIdType == types.EmPeerIdType_Uid {
			aBoxMsgApi.ReceiverId.PeerIdUnion = &gen_grpc.ChatPeerId_Uid{Uid: aConvMsg.ReceiverId.Uid}
		} else {
			aBoxMsgApi.ReceiverId.PeerIdUnion = &gen_grpc.ChatPeerId_GroupId{GroupId: aConvMsg.ReceiverId.GroupId}
		}
		aBoxMsgApi.Msg.MsgType = aConvMsg.Msg.MsgType
		aBoxMsgApi.Msg.SentTsMs = aConvMsg.Msg.SentTsMs
		aBoxMsgApi.Msg.SenderUid = aConvMsg.Msg.SenderUid
		aBoxMsgApi.Msg.MsgContent = aConvMsg.Msg.MsgContent
		aBoxMsgApi.Msg.ReadMsgId = aConvMsg.Msg.ReadMsgId
		res.MsgList = append(res.MsgList, aBoxMsgApi)
	}

	if len(msgList) > 0 {
		res.SeqId = msgList[len(msgList)-1].SeqId
	} else {
		res.SeqId = req.GetLocalSeqId()
	}

	res.ErrCode = gen_grpc.ErrCode_emErrCode_Ok
	return &res, nil
}
