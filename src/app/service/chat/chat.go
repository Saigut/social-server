package chat

import "errors"

/*
* msg
	- get_chat_msg
	- get_chat_msg_hist_with
	- send_chat_msg_to
*/

type ChatMsg struct {
	SenderUid string
	SentTsMs uint64
	MsgType uint16
	MsgContent string
}

type ChatMsgOfConv struct {
	ConvId uint64
	msg    ChatMsg
}

type ChatConvInfo struct {
	UidList []string
}

type GetChatMsgParam struct {
	Uid string
}

type GetChatMsgHistWithParam struct {
	ConvId uint64
}

type SendChatMsgToParam struct {
	ConvId uint64
	Msg    ChatMsg
}

type GetChatConvIdParam struct {
	Uid1 string
	Uid2 string
}


type Chat struct {

}

func (p *Chat) GetChatMsg() (msgs*ChatMsgOfConv, err error) {
	return nil, errors.New("method not implemented")
}

func (p *Chat) GetChatMsgHistWith() (msgs*ChatMsg, err error) {
	return nil, errors.New("method not implemented")
}

func (p *Chat) SendChatMsgTo(msg*ChatMsg) (err error) {
	return errors.New("method not implemented")
}

func (p *Chat) GetChatConvId(param*GetChatConvIdParam) (ConvId uint64, err error) {
	return 0, errors.New("method not implemented")
}
