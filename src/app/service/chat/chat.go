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
	MsgType uint32
	MsgContent string
}

type ChatMsgOfConv struct {
	ConvId uint64
	Msg    ChatMsg
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


type ChatT struct {

}

func (p *ChatT) GetChatMsg(uid string) (msgs *ChatMsgOfConv, err error) {
	return nil, errors.New("method not implemented")
}

func (p *ChatT) GetChatMsgHistWith(convId uint64) (msgs *ChatMsg, err error) {
	return nil, errors.New("method not implemented")
}

func (p *ChatT) SendChatMsgTo(msg *ChatMsg) (err error) {
	return errors.New("method not implemented")
}

func (p *ChatT) GetChatConvId(param *GetChatConvIdParam) (ConvId uint64, err error) {
	return 0, errors.New("method not implemented")
}
