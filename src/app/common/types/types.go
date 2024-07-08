package types

import gen_grpc "social_server/src/gen/grpc"

type SessId string

type SessCtx struct {
    SessId    SessId
    Uid       uint64
    Username  string
    CreatedAt uint64
    ExpiresAt uint64
}

type EmChatMsgType int32

const (
    EmChatMsgType_Text EmChatMsgType = 0

    EmChatMsgType_MarkRead EmChatMsgType = 50

    EmChatMsgType_ContactAddReq   EmChatMsgType = 100
    EmChatMsgType_ContactAdded    EmChatMsgType = 101
    EmChatMsgType_ContactRejected EmChatMsgType = 102
    EmChatMsgType_ContactDeleted  EmChatMsgType = 103

    EmChatMsgType_GroupCreated     EmChatMsgType = 201
    EmChatMsgType_GroupDeleted     EmChatMsgType = 202
    EmChatMsgType_GroupJoinReq     EmChatMsgType = 203
    EmChatMsgType_GroupUserJoined  EmChatMsgType = 204
    EmChatMsgType_GroupRejected    EmChatMsgType = 205
    EmChatMsgType_GroupUserLeft    EmChatMsgType = 206
    EmChatMsgType_GroupUserRemoved EmChatMsgType = 207
)

type EmPeerIdType uint
const (
    EmPeerIdType_Uid EmPeerIdType = 0
    EmPeerIdType_GroupId EmPeerIdType = 1
)

type PeerId struct {
    PeerIdType	  EmPeerIdType
    Uid       uint64
    GroupId   uint64
}

type ChatMsg struct {
    SenderUid  uint64
    SentTsMs   uint64
    MsgType    gen_grpc.ChatMsgType
    MsgContent string
    ReadMsgId  uint64
}

type ChatMsgOfConv struct {
    SeqId     uint64
    ConvMsgId uint64
    RandMsgId uint64
    ReceiverId PeerId
    Msg        ChatMsg
    IsRead     bool
    Status     uint32
}

type ChatConvInfo struct {
    UidList []uint64
}

type GetChatMsgParam struct {
    Uid uint64
}

type GetChatMsgHistWithParam struct {
    ConvId uint64
}

type SendMsgParam struct {
    PeerId PeerId
    Msg    ChatMsg
}

type GetChatConvIdParam struct {
    Uid1 uint64
    Uid2 uint64
}

type UmUserInfo struct {
    Uid      uint64
    Password string
    Username string
    Nickname string
    Email     string
    Avatar    string
}

type UmUserInfoValidateParam struct {
    Username  string
    Passphase string
    Email     string
}

type UmIsUsernameExistedParam struct {
    Username string
}

type UmUserAuthenticateParam struct {
    Username  string
    Passphase string
}

type UmRegisterParam struct {
    Username string `json:"username"`
    Passwd   string `json:"passwd"`
    Nickname string
    Email    string `json:"email"`
    Avatar   string
}

type UmUnregisterParam struct {
    Uid uint64 `json:"uid"`
}

type UmLoginParam struct {
    Uid    uint64 `json:"uid"`
    Passwd string `json:"passwd"`
}

type UmLogoutParam struct {
    Uid uint64 `json:"uid"`
}

type UmContactAddRequestParam struct {
    Uid        uint64
    ContactUid  uint64
}

type UmDelContactsParam struct {
    Uid        uint64
    ContactUid  uint64
}

type UmListContactsParam struct {
    Uid        uint64   `json:"uid"`
    ContactsUid []string `json:"contacts_uid"`
}

type UmGroupInfo struct {
    GroupId    uint64
    GroupName  string
    OwnerUid   uint64
    Avatar     string
    MemCount   uint64
    CreateTsMs uint64
}