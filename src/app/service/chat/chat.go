package chat

import (
	"errors"
	"fmt"
	"social_server/src/app/common/types"
	"social_server/src/app/data"
	. "social_server/src/utils/log"
	"sync"
)

type UserSync struct {
	mu   sync.Mutex
	cond *sync.Cond
}

type Chat struct {
	storage *data.DB
	cache *data.Cache

	// 用于通知新消息的channel。结构：map: uid -> ch
	//userChans map[uint64]chan struct{}

	userSyncs map[uint64]*UserSync
	rwMu      sync.RWMutex
}

func NewChat(storage *data.DB, cache *data.Cache) *Chat {
	return &Chat{
		storage: storage,
		cache: cache,
		//userChans: make(map[uint64]chan struct{}),
		userSyncs: make(map[uint64]*UserSync),
	}
}

func (c *Chat) getUserSync(uid uint64) *UserSync {
	// 尝试在无锁的情况下读取
	c.rwMu.RLock()
	us, exists := c.userSyncs[uid]
	if exists {
		return us
	}
	c.rwMu.RUnlock()

	// 如果不存在，再加锁进行写操作
	c.rwMu.Lock()
	defer c.rwMu.Unlock()

	// 双重检查
	us, exists = c.userSyncs[uid]
	if exists {
		return us
	}

	us = &UserSync{}
	us.cond = sync.NewCond(&us.mu)
	c.userSyncs[uid] = us

	return us
}

func (p *Chat) WaitForNewMessage(uid uint64) {
	us := p.getUserSync(uid)
	us.mu.Lock()
	defer us.mu.Unlock()
	us.cond.Wait()
}

func (p *Chat) GetChatMsgList(uid uint64, seqId uint64) (msgList []types.ChatMsgOfConv, err error) {

	msgList, err = p.storage.GetChatMsgList(uid, seqId)
	if err == nil {
		// 打印 msgList 长度
		Log.Debug("msgList len: %v", len(msgList))
		if len(msgList) > 0 {
			return msgList, nil
		}
	} else {
		if err.Error() != "no new msg" {
			return nil, fmt.Errorf("GetChatMsgList: %w", err)
		}
	}

	// 如果 seqId 是最新的，则等待
	Log.Debug("uid: %v", uid)
	p.WaitForNewMessage(uid)
	Log.Debug("new msg")

	msgList, err = p.storage.GetChatMsgList(uid, seqId)
	if err != nil {
		return nil, err
	}

	return msgList, nil
}

// 通知用户的所有 cond 停止等待
func (p *Chat) notifyAUserCond(uid uint64) {
	us := p.getUserSync(uid)
	us.mu.Lock()
	defer us.mu.Unlock()
	us.cond.Broadcast()
}

func (p *Chat) SendMsg(convMsg types.ChatMsgOfConv) (err error) {

	// 更新数据库
	err = p.storage.SendMsg(convMsg)
	if err != nil {
		return fmt.Errorf("storage.SendMsg: %w", err)
	}

	// 通知接收者和发送者
	if convMsg.ReceiverId.PeerIdType == types.EmPeerIdType_Uid {
		Log.Debug("notify user: %v", convMsg.ReceiverId.Uid)
		p.notifyAUserCond(convMsg.ReceiverId.Uid)
		Log.Debug("notify user: %v", convMsg.Msg.SenderUid)
		p.notifyAUserCond(convMsg.Msg.SenderUid)

	} else {
		// 获取群员列表
		memberList, err := p.storage.GetGroupMemberList(convMsg.ReceiverId.GroupId)
		if err != nil {
			return fmt.Errorf("storage.GetGroupMemberList: %w", err)
		}
		// 遍历群员列表并通知
		for _, uid := range memberList {
			Log.Debug("notify group member: %v", uid)
			p.notifyAUserCond(convMsg.ReceiverId.Uid)
		}
	}

	return nil
}

func (p *Chat) CreateGroupConv(uid uint64) (ConvId uint64, err error) {
	// 更新数据库
	ConvId, err = p.storage.CreateGroupConv(uid)
	if err != nil {
		return 0, err
	}
	return ConvId, nil
}

func (p *Chat) GetGroupConvList(uid uint64) (ConvIdList []uint64, err error) {
	// 从缓存中获取
	ConvIdList, err = p.cache.GetGroupConvList(uid)
	if err != nil {
		_, ok := err.(*data.CacheNotFoundError)
		if !ok {
			return nil, err
		}
	} else {
		return ConvIdList, nil
	}

	// 若无缓存，从数据库获取，并更新缓存
	ConvIdList, err = p.storage.GetGroupConvList(uid)
	if err != nil {
		return nil, err
	}

	err = p.cache.CacheGroupConvList(uid, ConvIdList)
	if err != nil {
		return nil, err
	}

	return nil, errors.New("method not implemented")
}

func (p *Chat) IsUserInConv(convId uint64, uid uint64) (inConv bool, err error) {
	// 检查缓存
	inConv, err = p.cache.IsUserInConv(convId, uid)
	if err != nil {
		_, ok := err.(*data.CacheNotFoundError)
		if !ok {
			return false, err
		}
	} else {
		return inConv, nil
	}

	// 若无缓存，从数据库获取，并更新缓存
	inConv, err = p.storage.IsUserInGroup(convId, uid)
	if err != nil {
		return false, err
	}

	err = p.cache.CacheIsUserInConv(convId, uid, inConv)
	if err != nil {
		return false, err
	}

	return inConv, nil
}
