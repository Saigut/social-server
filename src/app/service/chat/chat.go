package chat

import (
	"context"
	"fmt"
	"log"
	"os"
	"social_server/src/app/common/types"
	"social_server/src/app/data"
	gen_grpc "social_server/src/gen/grpc"
	. "social_server/src/utils/log"
	"strconv"
	"sync"
	"github.com/go-redis/redis/v8"
	"time"
)

type UserSync struct {
	condCh     chan struct{}
	channelName string
}

type Chat struct {
	storage *data.DB
	cache *data.Cache
	userSyncs map[uint64]*UserSync
	rwMu      sync.RWMutex
	redisClient *redis.Client
}

func NewChat(storage *data.DB, cache *data.Cache) *Chat {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := os.Getenv("REDIS_DB")

	if redisPort == "" {
		redisPort = "6379"
	}
	redisDBInt := 0
	if redisDB != "" {
		var err error
		redisDBInt, err = strconv.Atoi(redisDB)
		if err != nil {
			log.Fatalf("Invalid REDIS_DB value: %v", err)
		}
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPassword,
		DB:       redisDBInt,
	})

	// 尝试连接并发送 PING 命令
	var ctx = context.Background()
	err := redisClient.Ping(ctx).Err()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	} else {
		Log.Info("Connected to Redis successfully!")
	}

	return &Chat{
		storage: storage,
		cache: cache,
		//userChans: make(map[uint64]chan struct{}),
		userSyncs: make(map[uint64]*UserSync),
		redisClient: redisClient,
	}
}

func (p *Chat) getUserSync(uid uint64) (*UserSync) {
	p.rwMu.RLock()
	us, exists := p.userSyncs[uid]
	p.rwMu.RUnlock()
	if exists {
		return us
	}

	p.rwMu.Lock()
	defer p.rwMu.Unlock()

	// 双重检查
	us, exists = p.userSyncs[uid]
	if exists {
		return us
	}

	channelName := fmt.Sprintf("user_channel_%d", uid)
	us = &UserSync{}
	us = &UserSync{
		condCh:     make(chan struct{}),
		channelName: channelName,
	}
	p.userSyncs[uid] = us

	return us
}

func (p *Chat) waitForNewMessage(uid uint64) {
	us := p.getUserSync(uid)
	pubsub := p.redisClient.Subscribe(context.Background(), us.channelName)
	defer pubsub.Close()

	ch := pubsub.Channel()
	select {
	case <-ch:
		// 收到新消息通知
	case <-time.After(120 * time.Second):
		// 超时退出
	}
}

func (p *Chat) GetChatMsgList(uid uint64, seqId uint64) (msgList []types.ChatMsgOfConv, err error) {

	msgList, err = p.storage.ChatGetMsgList(uid, seqId)
	if err == nil {
		if len(msgList) > 0 {
			return msgList, nil
		}
	} else {
		if err.Error() != "no new msg" {
			return nil, fmt.Errorf("ChatGetMsgList: %w", err)
		}
	}

	// 如果 seqId 是最新的，则等待
	p.waitForNewMessage(uid)

	msgList, err = p.storage.ChatGetMsgList(uid, seqId)
	if err != nil {
		return nil, err
	}

	return msgList, nil
}

func (p *Chat) NotifyAUserCond(uid uint64) {
	us := p.getUserSync(uid)
	p.redisClient.Publish(context.Background(), us.channelName, "new message")
}

func (p *Chat) SendMsg(convMsg types.ChatMsgOfConv) (err error) {
	// 判断消息类型
	switch convMsg.Msg.MsgType {
	case gen_grpc.ChatMsgType_emChatMsgType_MarkRead:
		if convMsg.ReceiverId.PeerIdType == types.EmPeerIdType_Uid {
			err = p.storage.ChatMarkRead(convMsg.Msg.SenderUid, convMsg.ReceiverId.Uid, convMsg.Msg.ReadMsgId)
			if err != nil {
				return fmt.Errorf("ChatMarkRead: %w", err)
			}
		} else {
			err = p.storage.ChatReadGroupMsg(convMsg.Msg.SenderUid, convMsg.ReceiverId.GroupId, convMsg.Msg.ReadMsgId)
			if err != nil {
				return fmt.Errorf("ChatReadGroupMsg: %w", err)
			}
		}
	default:
		// do nothing
	}

	// 更新数据库
	err = p.storage.ChatSendMsg(convMsg)
	if err != nil {
		return fmt.Errorf("ChatSendMsg: %w", err)
	}

	// 通知接收者和发送者
	if convMsg.ReceiverId.PeerIdType == types.EmPeerIdType_Uid {
		p.NotifyAUserCond(convMsg.ReceiverId.Uid)
		p.NotifyAUserCond(convMsg.Msg.SenderUid)

	} else {
		// 获取群员列表
		memberList, err := p.storage.GroupGetMemList(convMsg.ReceiverId.GroupId)
		if err != nil {
			return fmt.Errorf("GroupGetMemList: %w", err)
		}
		// 遍历群员列表并通知
		for _, uid := range memberList {
			p.NotifyAUserCond(uid)
		}
	}

	return nil
}

func (p *Chat) SendMsgToUser(uid uint64, convMsg types.ChatMsgOfConv) (err error) {
	err = p.storage.ChatSendMsgToUser(uid, convMsg)
	if err != nil {
		return fmt.Errorf("ChatSendMsgToUser: %w", err)
	}
	p.NotifyAUserCond(uid)
	return nil
}

func (p *Chat) SendMsgToAdmins(convMsg types.ChatMsgOfConv) (err error) {
	err = p.storage.ChatSendMsgToAdmins(convMsg)
	if err != nil {
		return fmt.Errorf("ChatSendMsgToAdmins: %w", err)
	}
	// 获取管理员列表
	admins, err := p.storage.GroupGetAdminList(convMsg.ReceiverId.GroupId)
	if err != nil {
		return fmt.Errorf("GroupGetAdminList: %w", err)
	}
	for _, uid := range admins {
		p.NotifyAUserCond(uid)
	}
	return nil
}

func (p *Chat) AllocateGroupSeqId(groupId uint64) (seqId uint64, err error) {
	return p.storage.AllocateGroupSeqId(groupId)
}