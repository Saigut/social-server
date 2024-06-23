package data

import (
    "context"
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "github.com/go-redis/redis/v8"
    "log"
    "os"
    "social_server/src/app/common/types"
    . "social_server/src/utils/log"
    "strconv"
    "time"
)

type CacheUnknownError struct {
    Message string
}

func (e *CacheUnknownError) Error() string {
    return fmt.Sprintf("CacheUnknownError: %s", e.Message)
}

type CacheNotFoundError struct {
    Key string
}

func (e *CacheNotFoundError) Error() string {
    return fmt.Sprintf("CacheNotFoundError: Key %s not found", e.Key)
}

type Cache struct {
    client *redis.Client
}

func NewCache() *Cache {
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

    rdb := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
        Password: redisPassword,
        DB:       redisDBInt,
    })

    // 尝试连接并发送 PING 命令
    var ctx = context.Background()
    err := rdb.Ping(ctx).Err()
    if err != nil {
        Log.Debug("Failed to connect to Redis: %v", err)
    } else {
        Log.Debug("Connected to Redis successfully!")
    }

    return &Cache{client: rdb}
}

// Session
func GenerateSessionID(username string) (string, error) {
    // 获取当前时间戳
    timestamp := time.Now().UnixNano()

    // 生成随机字节
    const randomBytesLength = 16
    randomBytes := make([]byte, randomBytesLength)
    _, err := rand.Read(randomBytes)
    if err != nil {
        return "", fmt.Errorf("failed to generate random bytes: %w", err)
    }

    // 构建 session ID 原始数据
    data := fmt.Sprintf("%s:%d:%s", username, timestamp, hex.EncodeToString(randomBytes))

    // 计算 SHA-256 哈希值
    hash := sha256.Sum256([]byte(data))

    // 将哈希值编码为十六进制字符串
    sessionID := hex.EncodeToString(hash[:])

    return sessionID, nil
}

func (p *Cache) CreateSess(username string, uid uint64, timeoutTsS uint64) (sessId types.SessId, err error) {
    ctx := context.Background()

    var sessIdStr string
    sessIdStr, err = GenerateSessionID(username)
    if err != nil {
        return "", err
    }
    sessId = types.SessId(sessIdStr)

    createdAt := uint64(time.Now().Unix())
    expiresAt := createdAt + timeoutTsS

    // 会话上下文
    sessCtx := types.SessCtx{
        SessId:    sessId,
        Username:  username,
        Uid:       uid,
        CreatedAt: createdAt,
        ExpiresAt: expiresAt,
    }

    // 会话 key
    sessionKey := fmt.Sprintf("session:%s", sessIdStr)
    userSessionsKey := fmt.Sprintf("user:%s:sessions", username)

    // 使用事务创建会话并关联到用户
    _, err = p.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
        pipe.HSet(ctx, sessionKey, map[string]interface{}{
            "Uid":       sessCtx.Uid,
            "Username":  sessCtx.Username,
            "CreatedAt": sessCtx.CreatedAt,
            "ExpiresAt": sessCtx.ExpiresAt,
        })
        pipe.SAdd(ctx, userSessionsKey, sessIdStr)
        pipe.Expire(ctx, sessionKey, time.Duration(timeoutTsS)*time.Second) // 设置会话过期时间
        pipe.Expire(ctx, userSessionsKey, time.Duration(timeoutTsS)*time.Second) // 设置用户会话集合过期时间
        return nil
    })

    if err != nil {
        return "", err
    }

    return sessId, nil
}

func (p *Cache) GetSessCtx(sessId types.SessId) (sessCtx *types.SessCtx, err error) {
    // 查询会话
    sessionKey := fmt.Sprintf("session:%s", sessId)
    sessionData, err := p.client.HGetAll(context.Background(), sessionKey).Result()
    if err != nil {
        return nil, err
    }
    if len(sessionData) == 0 {
        return nil, fmt.Errorf("session not found")
    }

    uid, _ := strconv.ParseUint(sessionData["Uid"], 10, 64)
    createdAt, _ := strconv.ParseUint(sessionData["CreatedAt"], 10, 64)
    expiresAt, _ := strconv.ParseUint(sessionData["ExpiresAt"], 10, 64)
    username := sessionData["Username"]

    return &types.SessCtx{
        SessId:    sessId,
        Username:  username,
        Uid:       uid,
        CreatedAt: createdAt,
        ExpiresAt: expiresAt,
    }, nil
}

func (p *Cache) GetSessCtxByUsername(username string) (sessCtx *types.SessCtx, err error) {
    // 查找用户的会话集合 key
    userSessionsKey := fmt.Sprintf("user:%s:sessions", username)

    // 获取所有会话 ID
    sessIds, err := p.client.SMembers(context.Background(), userSessionsKey).Result()
    if err != nil {
        return nil, err
    }
    if len(sessIds) == 0 {
        return nil, fmt.Errorf("no sessions found for user: %s", username)
    }

    for _, sessId := range sessIds {
        sessCtx, err := p.GetSessCtx(types.SessId(sessId))
        if err != nil {
            // 如果会话不存在，继续处理下一个会话 ID
            if err.Error() == "session not found" {
                p.client.SRem(context.Background(), userSessionsKey, sessId)
                continue
            }
            return nil, err
        } else {
            return sessCtx, nil
        }
    }

    return nil, fmt.Errorf("no active sessions found for user: %s", username)
}

func (p *Cache) DeleteSess(sessId types.SessId) (err error) {
    // 查找会话 key
    sessionKey := fmt.Sprintf("session:%s", sessId)

    // 获取会话数据
    sessionData, err := p.client.HGetAll(context.Background(), sessionKey).Result()
    if err != nil {
        return err
    }
    if len(sessionData) == 0 {
        return fmt.Errorf("session not found")
    }

    // 获取用户名
    username := sessionData["Username"]
    userSessionsKey := fmt.Sprintf("user:%s:sessions", username)

    // 删除会话和用户会话集合中的会话 ID
    _, err = p.client.TxPipelined(context.Background(), func(pipe redis.Pipeliner) error {
        pipe.Del(context.Background(), sessionKey)
        pipe.SRem(context.Background(), userSessionsKey, sessId)
        return nil
    })

    if err != nil {
        return err
    }

    return nil
}


// Chat
func (p *Cache) GetChatMsgList(uid uint64, seqId uint64) (msgs []types.ChatMsgOfConv, err error) {
    ctx := context.Background()
    key := fmt.Sprintf("chat:msglist:%d", uid)
    result, err := p.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return nil, &CacheNotFoundError{Key: key}
    } else if err != nil {
        return nil, err
    }

    // todo seqId

    // 反序列化
    err = json.Unmarshal([]byte(result), &msgs)
    if err != nil {
        return nil, err
    }

    return msgs, nil
}

func (p *Cache) CacheChatMsgList(uid uint64, seqId uint64, msgs []types.ChatMsgOfConv) (err error) {
    ctx := context.Background()
    key := fmt.Sprintf("chat:msglist:%d", uid)
    data, err := json.Marshal(msgs)
    if err != nil {
        return err
    }
    return p.client.Set(ctx, key, data, 24*time.Hour).Err()
}

func (p *Cache) ClearCacheChatMsgList(uid uint64, seqId uint64) (err error) {
    ctx := context.Background()
    key := fmt.Sprintf("chat:msglist:%d", uid)
    return p.client.Del(ctx, key).Err()
}

func (p *Cache) GetConvMsgHist(peerId types.PeerId) (msgs []types.ChatMsg, err error) {
    ctx := context.Background()
    key := fmt.Sprintf("conv:msghist:%s", peerId)
    result, err := p.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return nil, &CacheNotFoundError{Key: key}
    } else if err != nil {
        return nil, err
    }

    // 反序列化
    err = json.Unmarshal([]byte(result), &msgs)
    if err != nil {
        return nil, err
    }

    return msgs, nil
}

func (p *Cache) CacheConvMsgHist(peerId types.PeerId, msgs []types.ChatMsg) (err error) {
    ctx := context.Background()
    key := fmt.Sprintf("conv:msghist:%s", peerId)
    data, err := json.Marshal(msgs)
    if err != nil {
        return err
    }
    return p.client.Set(ctx, key, data, 24*time.Hour).Err()
}

func (p *Cache) ClearCacheConvMsgHist(peerId types.PeerId) (err error) {
    ctx := context.Background()
    key := fmt.Sprintf("conv:msghist:%s", peerId)
    return p.client.Del(ctx, key).Err()
}

func (p *Cache) SendMsg(peerId types.PeerId, msg types.ChatMsg) (err error) {
    // 具体实现根据业务逻辑
    return errors.New("method not implemented")
}

func (p *Cache) CreateGroupConv(uid uint64) (ConvId uint64, err error) {
    // 具体实现根据业务逻辑
    return 0, errors.New("method not implemented")
}

func (p *Cache) GetGroupConvList(uid uint64) (ConvIdList []uint64, err error) {
    ctx := context.Background()
    key := fmt.Sprintf("group:convlist:%d", uid)
    result, err := p.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return nil, &CacheNotFoundError{Key: key}
    } else if err != nil {
        return nil, err
    }

    // 反序列化
    err = json.Unmarshal([]byte(result), &ConvIdList)
    if err != nil {
        return nil, err
    }

    return ConvIdList, nil
}

func (p *Cache) CacheGroupConvList(uid uint64, ConvIdList []uint64) (err error) {
    ctx := context.Background()
    key := fmt.Sprintf("group:convlist:%d", uid)
    data, err := json.Marshal(ConvIdList)
    if err != nil {
        return err
    }
    return p.client.Set(ctx, key, data, 24*time.Hour).Err()
}

func (p *Cache) ClearCacheGroupConvList(uid uint64) (err error) {
    ctx := context.Background()
    key := fmt.Sprintf("group:convlist:%d", uid)
    return p.client.Del(ctx, key).Err()
}

func (p *Cache) IsUserInConv(convId uint64, uid uint64) (inConv bool, err error) {
    ctx := context.Background()
    key := fmt.Sprintf("conv:isuserin:%d:%d", convId, uid)
    result, err := p.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return false, &CacheNotFoundError{Key: key}
    } else if err != nil {
        return false, err
    }

    inConv = result == "true"
    return inConv, nil
}

func (p *Cache) CacheIsUserInConv(convId uint64, uid uint64, inConv bool) (err error) {
    ctx := context.Background()
    key := fmt.Sprintf("conv:isuserin:%d:%d", convId, uid)
    return p.client.Set(ctx, key, inConv, 24*time.Hour).Err()
}

func (p *Cache) ClearCacheIsUserInConv(convId uint64, uid uint64) (err error) {
    ctx := context.Background()
    key := fmt.Sprintf("conv:isuserin:%d:%d", convId, uid)
    return p.client.Del(ctx, key).Err()
}

// UserMgmt
func (p *Cache) IsUsernameExisted(userName string) (isExisted bool, err error) {
    ctx := context.Background()
    key := fmt.Sprintf("user:username:%s", userName)
    result, err := p.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return false, &CacheNotFoundError{Key: key}
    } else if err != nil {
        return false, err
    }

    isExisted = result == "true"
    return isExisted, nil
}

func (p *Cache) CacheIsUsernameExisted(userName string, isExisted bool) (err error) {
    ctx := context.Background()
    key := fmt.Sprintf("user:username:%s", userName)
    return p.client.Set(ctx, key, isExisted, 24*time.Hour).Err()
}

func (p *Cache) ClearCacheIsUsernameExisted(userName string) (err error) {
    ctx := context.Background()
    key := fmt.Sprintf("user:username:%s", userName)
    return p.client.Del(ctx, key).Err()
}

func (p *Cache) UserAuthenticate(param *types.UmUserAuthenticateParam) (pass bool, err error) {
    ctx := context.Background()
    // 查询用户信息
    key := fmt.Sprintf("user:userinfo:%s", param.Username)
    result, err := p.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return false, &CacheNotFoundError{Key: key}
    } else if err != nil {
        return false, err
    } else {
        // 反序列化
        var user types.UmUserInfo
        err = json.Unmarshal([]byte(result), &user)
        if err != nil {
            return false, err
        }

        // 验证密码
        if user.Passphase == param.Passphase {
            return true, nil
        } else {
            return false, nil
        }
    }
}

func (p *Cache) CacheUserAuthenticate(user *types.UmUserInfo) (err error) {
    ctx := context.Background()
    key := fmt.Sprintf("user:userinfo:%s", user.Username)
    data, err := json.Marshal(user)
    if err != nil {
        return err
    }
    return p.client.Set(ctx, key, data, 24*time.Hour).Err()
}

func (p *Cache) ClearCacheUserAuthenticate(username string) (err error) {
    ctx := context.Background()
    key := fmt.Sprintf("user:userinfo:%s", username)
    return p.client.Del(ctx, key).Err()
}

func (p *Cache) Register(param *types.UmRegisterParam) (err error) {
    // 具体实现根据业务逻辑
    return errors.New("method not implemented")
}

func (p *Cache) Unregister(param *types.UmUnregisterParam) (err error) {
    // 具体实现根据业务逻辑
    return errors.New("method not implemented")
}

func (p *Cache) AddFriends(param *types.UmAddFriendsParam) (err error) {
    // 具体实现根据业务逻辑
    return errors.New("method not implemented")
}

func (p *Cache) DelFriends(param *types.UmDelFriendsParam) (err error) {
    // 具体实现根据业务逻辑
    return errors.New("method not implemented")
}

func (p *Cache) GetFriendList(uid uint64) (friendsUid []uint64, err error) {
    ctx := context.Background()
    key := fmt.Sprintf("friend:list:%d", uid)
    result, err := p.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return nil, &CacheNotFoundError{Key: key}
    } else if err != nil {
        return nil, err
    }

    // 反序列化
    err = json.Unmarshal([]byte(result), &friendsUid)
    if err != nil {
        return nil, err
    }

    return friendsUid, nil
}

func (p *Cache) CacheFriendList(uid uint64, friendsUid []uint64) (err error) {
    ctx := context.Background()
    key := fmt.Sprintf("friend:list:%d", uid)
    data, err := json.Marshal(friendsUid)
    if err != nil {
        return err
    }
    return p.client.Set(ctx, key, data, 24*time.Hour).Err()
}

func (p *Cache) ClearCacheFriendList(uid uint64) (err error) {
    ctx := context.Background()
    key := fmt.Sprintf("friend:list:%d", uid)
    return p.client.Del(ctx, key).Err()
}