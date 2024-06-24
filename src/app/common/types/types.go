package types

type SessId string

type SessCtx struct {
    SessId    SessId
    Username  string
    Uid       uint64
    CreatedAt uint64
    ExpiresAt uint64
}

type EmChatMsgType int32

const (
    EmChatMsgType_Text EmChatMsgType = 0
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
    MsgType    EmChatMsgType
    MsgContent string
}

type ChatMsgOfConv struct {
    SeqId      uint64
    MsgId      uint64
    RandMsgId  uint64
    ReceiverId PeerId
    Msg        ChatMsg
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
    Uid       uint64
    Username  string
    Passphase string
    Email     string
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
    Email    string `json:"email"`
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

type UmAddFriendsParam struct {
    Uid        uint64
    FriendUid  uint64
}

type UmDelFriendsParam struct {
    Uid        uint64
    FriendUid  uint64
}

type UmListFriendsParam struct {
    Uid        uint64   `json:"uid"`
    FriendsUid []string `json:"friends_uid"`
}

