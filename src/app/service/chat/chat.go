package chat

import "errors"

/*
* msg
	- get_chat_msg
	- get_chat_msg_hist_with
	- send_chat_msg_to
*/

type EmChatMsgType int32

const (
	EmChatMsgType_emChatMsgType_Text EmChatMsgType = 0
)

type PeerId struct {
	PeerIdType		uint  // 0 peerUid, 1 groupConvId
	PeerUid       uint64
	GroupConvId   uint64
}

type ChatMsg struct {
	SenderUid uint64
	SentTsMs uint64
	MsgType EmChatMsgType
	MsgContent string
}

type ChatMsgOfConv struct {
	PeerId PeerId
	Msg    ChatMsg
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

type SendChatMsgToParam struct {
	ConvId uint64
	Msg    ChatMsg
}

type GetChatConvIdParam struct {
	Uid1 uint64
	Uid2 uint64
}


type ChatT struct {

}

func (p *ChatT) GetChatMsgList(uid uint64) (msgs []ChatMsgOfConv, err error) {
	return nil, errors.New("method not implemented")
}

func (p *ChatT) GetConvMsgHist(peerId PeerId) (msgs []ChatMsg, err error) {
	return nil, errors.New("method not implemented")
}

func (p *ChatT) SendChatMsgTo(peerId PeerId, msg ChatMsg) (err error) {
	return errors.New("method not implemented")
}

func (p *ChatT) CreateGroupConv(uid uint64) (ConvId uint64, err error) {
	return 0, errors.New("method not implemented")
}

func (p *ChatT) GetGroupConvList(uid uint64) (ConvIdList []uint64, err error) {
	return nil, errors.New("method not implemented")
}

func (p *ChatT) IsUserInConv(convId uint64, uid uint64) (inConv bool) {
	return false
}
