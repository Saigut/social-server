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

    return &Cache{client: redisClient}
}

func GenerateSessionID(uid uint64) (string, error) {
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
    data := fmt.Sprintf("%d:%d:%s", uid, timestamp, hex.EncodeToString(randomBytes))

    // 计算 SHA-256 哈希值
    hash := sha256.Sum256([]byte(data))

    // 将哈希值编码为十六进制字符串
    sessionID := hex.EncodeToString(hash[:])

    return sessionID, nil
}

func (p *Cache) CreateSess(username string, uid uint64, expireAfterSecs uint64) (sessId types.SessId, err error) {
    ctx := context.Background()

    var sessIdStr string
    sessIdStr, err = GenerateSessionID(uid)
    if err != nil {
        return "", err
    }

    sessionKey := fmt.Sprintf("session:%s", sessIdStr)
    exists, err := p.client.Exists(ctx, sessionKey).Result()
    if err != nil {
        return "", fmt.Errorf("failed to check session existence: %w", err)
    }
    if exists > 0 {
        return "", fmt.Errorf("session ID already exists")
    }

    sessId = types.SessId(sessIdStr)

    createdAt := uint64(time.Now().Unix())
    expiresAt := createdAt + expireAfterSecs

    // 会话上下文
    sessCtx := types.SessCtx{
        SessId:    sessId,
        Username:  username,
        Uid:       uid,
        CreatedAt: createdAt,
        ExpiresAt: expiresAt,
    }

    userSessionsKey := fmt.Sprintf("user:%v:sessions", uid)

    // 使用事务创建会话并关联到用户
    _, err = p.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
        pipe.HSet(ctx, sessionKey, map[string]interface{}{
            "Uid":       sessCtx.Uid,
            "Username":  sessCtx.Username,
            "CreatedAt": sessCtx.CreatedAt,
            "ExpiresAt": sessCtx.ExpiresAt,
        })
        pipe.SAdd(ctx, userSessionsKey, sessIdStr)
        pipe.Expire(ctx, sessionKey, time.Duration(expireAfterSecs)*time.Second)      // 设置会话过期时间
        pipe.Expire(ctx, userSessionsKey, time.Duration(expireAfterSecs)*time.Second) // 设置用户会话集合过期时间
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

func (p *Cache) RenewSessCtx(sessCtx *types.SessCtx, expireAfterSecs uint64) (err error) {
    if expireAfterSecs == 0 {
        return fmt.Errorf("expireAfterSecs must be greater than zero")
    }

    ctx := context.Background()

    // 计算新的过期时间
    currentTime := uint64(time.Now().Unix())
    newExpiresAt := currentTime + expireAfterSecs

    // 获取会话键和用户会话键
    sessionKey := fmt.Sprintf("session:%s", sessCtx.SessId)
    userSessionsKey := fmt.Sprintf("user:%v:sessions", sessCtx.Uid)

    // 使用事务来更新会话信息和过期时间
    _, err = p.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
        // 更新会话上下文中的过期时间
        if err := pipe.HSet(ctx, sessionKey, "ExpiresAt", newExpiresAt).Err(); err != nil {
            return fmt.Errorf("failed to set new expiration in session context: %w", err)
        }
        // 续期会话键
        if err := pipe.Expire(ctx, sessionKey, time.Duration(expireAfterSecs)*time.Second).Err(); err != nil {
            return fmt.Errorf("failed to expire session key: %w", err)
        }
        // 续期用户会话集合键
        if err := pipe.Expire(ctx, userSessionsKey, time.Duration(expireAfterSecs)*time.Second).Err(); err != nil {
            return fmt.Errorf("failed to expire user sessions key: %w", err)
        }
        return nil
    })

    if err != nil {
        return fmt.Errorf("TxPipelined: %w", err)
    }

    return nil
}

func (p *Cache) RenewSessCtxBySessId(sessId types.SessId, expireAfterSecs uint64) error {
    // 获取会话上下文
    sessCtx, err := p.GetSessCtx(sessId)
    if err != nil {
        return fmt.Errorf("GetSessCtx: %v", err)
    }
    if sessCtx == nil {
        return fmt.Errorf("session not found")
    }
    return p.RenewSessCtx(sessCtx, expireAfterSecs)
}

func (p *Cache) GetSessCtxByUid(uid uint64) (sessCtx *types.SessCtx, err error) {
    // 查找用户的会话集合 key
    userSessionsKey := fmt.Sprintf("user:%v:sessions", uid)

    // 获取所有会话 ID
    sessIds, err := p.client.SMembers(context.Background(), userSessionsKey).Result()
    if err != nil {
        return nil, err
    }
    if len(sessIds) == 0 {
        return nil, fmt.Errorf("no sessions found for user: %s", uid)
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

    return nil, fmt.Errorf("no active sessions found for user: %s", uid)
}

func (p *Cache) DeleteSess(sessId types.SessId) (err error) {
    // 查找会话 key
    sessionKey := fmt.Sprintf("session:%s", sessId)

    // 获取会话数据
    sessionData, err := p.client.HGetAll(context.Background(), sessionKey).Result()
    if err != nil {
        return fmt.Errorf("HGetAll: %w", err)
    }
    if len(sessionData) == 0 {
        return fmt.Errorf("session not found")
    }

    // 获取uid
    uid := sessionData["Uid"]
    userSessionsKey := fmt.Sprintf("user:%v:sessions", uid)

    // 删除会话和用户会话集合中的会话 ID
    _, err = p.client.TxPipelined(context.Background(), func(pipe redis.Pipeliner) error {
        pipe.Del(context.Background(), sessionKey)
        pipe.SRem(context.Background(), userSessionsKey, string(sessId))
        return nil
    })

    if err != nil {
        return fmt.Errorf("TxPipelined: %w", err)
    }

    return nil
}

func (p *Cache) DeleteUserSess(uid uint64) error {
    ctx := context.Background()
    userSessionsKey := fmt.Sprintf("user:%v:sessions", uid)

    // 获取用户的所有会话ID
    sessIds, err := p.client.SMembers(ctx, userSessionsKey).Result()
    if err != nil {
        return fmt.Errorf("SMembers: %w", err)
    }

    if len(sessIds) == 0 {
        return fmt.Errorf("no sessions found for user %d", uid)
    }

    // 使用事务删除用户的所有会话
    _, err = p.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
        for _, sessIdStr := range sessIds {
            sessionKey := fmt.Sprintf("session:%s", sessIdStr)
            pipe.Del(ctx, sessionKey) // 删除会话数据
        }
        pipe.Del(ctx, userSessionsKey) // 删除用户会话集合
        return nil
    })

    if err != nil {
        return fmt.Errorf("TxPipelined: %w", err)
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

func (p *Cache) SendMsg(peerId types.PeerId, msg types.ChatMsg) (err error) {
    // 具体实现根据业务逻辑
    return errors.New("method not implemented")
}

func (p *Cache) CreateGroupConv(uid uint64) (ConvId uint64, err error) {
    // 具体实现根据业务逻辑
    return 0, errors.New("method not implemented")
}

func (p *Cache) GroupGetList(uid uint64) (ConvIdList []uint64, err error) {
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
        if user.Password == param.Passphase {
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

func (p *Cache) AddContacts(param *types.UmContactAddRequestParam) (err error) {
    // 具体实现根据业务逻辑
    return errors.New("method not implemented")
}

func (p *Cache) DelContacts(param *types.UmDelContactsParam) (err error) {
    // 具体实现根据业务逻辑
    return errors.New("method not implemented")
}

func (p *Cache) GetContactList(uid uint64) (contactsUid []uint64, err error) {
    ctx := context.Background()
    key := fmt.Sprintf("contact:list:%d", uid)
    result, err := p.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return nil, &CacheNotFoundError{Key: key}
    } else if err != nil {
        return nil, err
    }

    // 反序列化
    err = json.Unmarshal([]byte(result), &contactsUid)
    if err != nil {
        return nil, err
    }

    return contactsUid, nil
}

func (p *Cache) CacheContactList(uid uint64, contactsUid []uint64) (err error) {
    ctx := context.Background()
    key := fmt.Sprintf("contact:list:%d", uid)
    data, err := json.Marshal(contactsUid)
    if err != nil {
        return err
    }
    return p.client.Set(ctx, key, data, 24*time.Hour).Err()
}

func (p *Cache) ClearCacheContactList(uid uint64) (err error) {
    ctx := context.Background()
    key := fmt.Sprintf("contact:list:%d", uid)
    return p.client.Del(ctx, key).Err()
}