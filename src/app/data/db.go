package data

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"social_server/src/app/common/types"
	. "social_server/src/utils/log"
	"sync"
	"time"
)

// Users table structure
type User struct {
	UserID    int
	Username  string
	Password  string
	Email     string
	CreatedAt time.Time
}

// Friends table structure
type Friend struct {
	UserID    int
	FriendID  int
	Status    int
	CreatedAt time.Time
}

// ChatGroups table structure
type ChatGroup struct {
	GroupID   int
	GroupName string
	OwnerID   int
	CreatedAt time.Time
}

// ChatGroupMembers table structure
type ChatGroupMember struct {
	GroupID  int
	UserID   int
	Role     string
	JoinedAt time.Time
}

// Messages table structure
type Message struct {
	MessageID   int
	SenderID    int
	ReceiverID  int
	GroupID     int
	Content     string
	MessageType int
	SentAt      time.Time
}

type InboxMsg struct {
	UserID      int
	SeqID       int
	SenderID    int
	ReceiverID  sql.NullInt64
	GroupID     sql.NullInt64
	Content     string
	MessageType int
	SentAt      time.Time
}

func tryPush(ch chan struct{}) bool {
	select {
	case ch <- struct{}{}:
		return true
	default:
		return false
	}
}

type DB struct {
	dsn                  string
	maxReconnectInterval time.Duration
	connectTimeout       time.Duration
	maxPendingRequests   int

	db                   *sql.DB
	dbMutex              sync.Mutex
	connectOnce          sync.Once
	reconnectChan        chan struct{}
	pendingRequestsCount int

	retrying 	     	  bool
}

func NewStorage() *DB {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	if dbPort == "" {
		dbPort = "3306"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbHost, dbPort, dbName)

	return &DB{
		dsn:                  dsn,
		maxReconnectInterval: 10 * time.Second,
		connectTimeout:       30 * time.Second,
		maxPendingRequests:   10000,
		reconnectChan:        make(chan struct{}, 1),
		retrying:			  false,
	}
}

func (p *DB) connectDb() error {
	var err error
	p.db, err = sql.Open("mysql", p.dsn)
	if err != nil {
		return err
	}
	p.db.SetConnMaxLifetime(p.connectTimeout)
	p.db.SetMaxOpenConns(10)
	p.db.SetMaxIdleConns(10)
	return p.db.Ping()
}

func (p *DB) reconnectTask() {
	var err error
	for {
		<-p.reconnectChan

		// 检查连接状态
		if p.db != nil {
			err = p.db.Ping()
			if err == nil {
				continue
			} else {
				p.db.Close()
				p.db = nil
			}
		}

		// 连接数据库
		p.retrying = true
		for {
			err = p.connectDb()
			if err == nil {
				Log.Info("Connected to the database successfully.")
				p.retrying = false
				break
			}

			Log.Error("Failed to reconnect: %v. retrying in 10 seconds...", err.Error())
			time.Sleep(p.maxReconnectInterval)
		}
	}
}

func (p *DB) triggerReconnect() {
	tryPush(p.reconnectChan)
}

func (p *DB) Init() {
	p.connectOnce.Do(func() {
		go p.reconnectTask()
		tryPush(p.reconnectChan)
	})
}

func (p *DB) withReconnectHandling(action func() (interface{}, error)) (interface{}, error)  {
	if p.db == nil {
		return nil, errors.New("db not connect yet")
	}
	if p.retrying {
		return nil, errors.New("db connection issue")
	}
	if p.pendingRequestsCount > p.maxPendingRequests {
		return nil, errors.New("too many pending db requests")
	}

	p.dbMutex.Lock()
	p.pendingRequestsCount++
	p.dbMutex.Unlock()
	defer func() {
		p.dbMutex.Lock()
		p.pendingRequestsCount--
		p.dbMutex.Unlock()
	}()

	ret, err := action()
	if err != nil {
		p.triggerReconnect()
		return nil, fmt.Errorf("action: %w", err)
	}

	return ret, nil
}

func (p *DB) sqlExec(sqlStr string, args ...interface{}) (sql.Result, error) {
	ret, err := p.withReconnectHandling(func() (interface{}, error) {
		return p.db.Exec(sqlStr, args...)
	})
	if err != nil {
		return nil, fmt.Errorf("withReconnectHandling: %w", err)
	}
	return ret.(sql.Result), nil
}

func (p *DB) queryRow(query string, args ...interface{}) (*sql.Row, error) {
	ret, err := p.withReconnectHandling(func() (interface{}, error) {
		return p.db.QueryRow(query, args...), nil
	})
	if err != nil {
		return nil, fmt.Errorf("withReconnectHandling: %w", err)
	}
	return ret.(*sql.Row), nil
}

func (p *DB) queryRows(query string, args ...interface{}) (*sql.Rows, error) {
	ret, err := p.withReconnectHandling(func() (interface{}, error) {
		return p.db.Query(query, args...)
	})
	if err != nil {
		return nil, err
	}
	return ret.(*sql.Rows), nil
}

func convertDbMsgToChatMsg(rowMsg InboxMsg) (msg types.ChatMsg, err error) {
	msg.SenderUid = uint64(rowMsg.SenderID)
	msg.SentTsMs = uint64(rowMsg.SentAt.UnixNano() / 1e6)
	msg.MsgContent = rowMsg.Content
	switch rowMsg.MessageType {
	case 0:
		msg.MsgType = types.EmChatMsgType_Text
	default:
		return msg, errors.New("unknown message type")
	}
	return msg, nil
}

func convertDbMsgToChatMsgOfConv(rowMsg InboxMsg) (msg types.ChatMsgOfConv, err error) {
	msg.Msg, err = convertDbMsgToChatMsg(rowMsg)
	if err != nil {
		return msg, fmt.Errorf("convertDbMsgToChatMsg: %w", err)
	}

	msg.SeqId = uint64(rowMsg.SeqID)

	if rowMsg.GroupID.Valid {
		msg.PeerId.PeerIdType = types.EmPeerIdType_GroupConvId
		msg.PeerId.GroupConvId = uint64(rowMsg.GroupID.Int64)
	} else if rowMsg.ReceiverID.Valid {
		msg.PeerId.PeerIdType = types.EmPeerIdType_PeerUid
		msg.PeerId.PeerUid = uint64(rowMsg.SenderID)
	} else {
		return msg, errors.New("invalid peer id")
	}

	return msg, nil
}


// Chat
func (p *DB) GetChatMsgList(uid uint64, seqId uint64) (msgs []types.ChatMsgOfConv, err error) {
	rows, err := p.queryRows(`
        SELECT user_id, seq_id, sender_id, receiver_id, group_id, content, message_type, sent_at 
        FROM Inbox
        WHERE user_id = ? AND seq_id > ?
    `, uid, seqId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no new msg")
		} else {
			return nil, fmt.Errorf("queryRows: %w", err)
		}
	}
	defer rows.Close()

	for rows.Next() {
		var msg types.ChatMsgOfConv
		var rowMsg InboxMsg
		err = rows.Scan(&rowMsg.UserID, &rowMsg.SeqID, &rowMsg.SenderID, &rowMsg.ReceiverID, &rowMsg.GroupID, &rowMsg.Content, &rowMsg.MessageType, &rowMsg.SentAt)
		Log.Debug("rowMsg: %+v", rowMsg)
		msg, err = convertDbMsgToChatMsgOfConv(rowMsg)
		if err != nil {
			return nil, err
		}
		Log.Debug("msg: %+v", msg)
		msgs = append(msgs, msg)
	}

	return msgs, nil
}

func (p *DB) GetConvMsgHist(peerId types.PeerId) (msgs []types.ChatMsg, err error) {
	rows, err := p.queryRows("SELECT * FROM Messages WHERE (sender_id = ? OR receiver_id = ?) AND message_type = 'private'", peerId, peerId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var msg types.ChatMsg
		var rowMsg Message
		err := rows.Scan(&rowMsg.MessageID, &rowMsg.SenderID, &rowMsg.ReceiverID, &rowMsg.GroupID, &rowMsg.Content, &rowMsg.MessageType, &rowMsg.SentAt)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}

	return msgs, nil
}

func (p *DB) SendMsg(peerId types.PeerId, msg types.ChatMsg) (err error) {

	if peerId.PeerIdType == types.EmPeerIdType_PeerUid {
		// 分配 seqId
		var seqId uint64
		seqId, err = p.AllocateSeqId(peerId.PeerUid)
		if err != nil {
			return fmt.Errorf("AllocateSeqId: %w", err)
		}

		// 添加消息
		_, err = p.sqlExec("INSERT INTO Inbox (user_id, seq_id, sender_id, receiver_id, group_id, content, message_type) VALUES (?, ?, ?, ?, ?, ?, ?)",
			peerId.PeerUid, seqId, msg.SenderUid, peerId.PeerUid, nil, msg.MsgContent, msg.MsgType)
		if err != nil {
			return fmt.Errorf("user sqlExec: %w", err)
		}

	} else {
		// 获取群员群员列表
		memberList, err := p.GetGroupMemberList(peerId.GroupConvId)
		if err != nil {
			return fmt.Errorf("GetGroupMemberList: %w", err)
		}

		// todo: 事务
		for _, memberUid := range memberList {
			// 分配 seqId
			seqId, err := p.AllocateSeqId(memberUid)
			if err != nil {
				return fmt.Errorf("AllocateSeqId: %w", err)
			}

			// 添加消息
			_, err = p.sqlExec("INSERT INTO Inbox (user_id, seq_id, sender_id, receiver_id, group_id, content, message_type) VALUES (?, ?, ?, ?, ?, ?, ?)",
				memberUid, seqId, msg.SenderUid, nil, peerId.GroupConvId, msg.MsgContent, msg.MsgType)
			if err != nil {
				return fmt.Errorf("group sqlExec: %w", err)
			}
		}
	}

	return err
}

func (p *DB) CreateGroupConv(uid uint64) (ConvId uint64, err error) {
	res, err := p.sqlExec("INSERT INTO ChatGroups (group_name, owner_id) VALUES (?, ?)", "", uid)
	if err != nil {
		return 0, err
	}

	groupID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	_, err = p.sqlExec("INSERT INTO ChatGroupMembers (group_id, user_id, role) VALUES (?, ?, 'admin')", groupID, uid)
	return uint64(groupID), err
}

func (p *DB) GetGroupConvList(uid uint64) (ConvIdList []uint64, err error) {
	rows, err := p.queryRows("SELECT group_id FROM ChatGroupMembers WHERE user_id = ?", uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var groupID uint64
		err := rows.Scan(&groupID)
		if err != nil {
			return nil, err
		}
		ConvIdList = append(ConvIdList, groupID)
	}

	return ConvIdList, nil
}

// 获取群组成员列表
func (p *DB) GetGroupMemberList(groupId uint64) (membersUid []uint64, err error) {
	rows, err := p.queryRows("SELECT user_id FROM ChatGroupMembers WHERE group_id = ?", groupId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var uid uint64
		err := rows.Scan(&uid)
		if err != nil {
			return nil, err
		}
		membersUid = append(membersUid, uid)
	}
	return membersUid, nil
}

func (p *DB) IsUserInGroup(groupId uint64, uid uint64) (inGroup bool, err error) {
	row, err := p.queryRow("SELECT COUNT(*) FROM ChatGroupMembers WHERE group_id = ? AND user_id = ?", groupId, uid)
	var count int
	err = row.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UserMgmt
func (p *DB) IsUsernameExisted(userName string) (bool, error) {
	row, err := p.queryRow("SELECT COUNT(*) FROM Users WHERE username = ?", userName)
	var count int
	err = row.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// 获取用户信息
func (p *DB) GetUserInfo(uid uint64) (user *types.UmUserInfo, err error) {
	row, err := p.queryRow("SELECT user_id, username, email FROM Users WHERE user_id = ?", uid)
	if err != nil {
		return nil, err
	}
	var userRow User
	err = row.Scan(&userRow.UserID, &userRow.Username, &userRow.Email)
	if err != nil {
		return nil, err
	}
	return &types.UmUserInfo{
		Passphase: userRow.Password,
		Email:     userRow.Email,
		Uid:       uint64(userRow.UserID),
		Username:  userRow.Username,
	}, nil
}

func (p *DB) GetUserInfoByUsername(username string) (user *types.UmUserInfo, err error) {
	row, err := p.queryRow("SELECT user_id, username, email FROM Users WHERE username = ?", username)
	if err != nil {
		return nil, err
	}
	var userRow User
	err = row.Scan(&userRow.UserID, &userRow.Username, &userRow.Email)
	if err != nil {
		return nil, err
	}
	return &types.UmUserInfo{
		Passphase: userRow.Password,
		Email:     userRow.Email,
		Uid:       uint64(userRow.UserID),
		Username:  userRow.Username,
	}, nil
}

func (p *DB) UserAuthenticate(param *types.UmUserAuthenticateParam) (bool, error) {
	row, err := p.queryRow("SELECT COUNT(*) FROM Users WHERE username = ? AND password = ?", param.Username, param.Passphase)
	var count int
	err = row.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (p *DB) Register(param *types.UmRegisterParam) (error) {
	_, err := p.sqlExec("INSERT INTO Users (username, password, email) VALUES (?, ?, ?)", param.Username, param.Passwd, param.Email)
	return err
}

func (p *DB) Unregister(param *types.UmUnregisterParam) (error) {
	_, err := p.sqlExec("DELETE FROM Users WHERE user_id = ?", param.Uid)
	return err
}

func (p *DB) AddFriends(param *types.UmAddFriendsParam) (error) {
	_, err := p.sqlExec("INSERT INTO Friends (user_id, friend_id, status) VALUES (?, ?, 1)", param.Uid, param.FriendUid)
	return err
}

func (p *DB) DelFriend(param *types.UmDelFriendsParam) (error) {
	_, err := p.sqlExec("DELETE FROM Friends WHERE user_id = ? AND friend_id = ?", param.Uid, param.FriendUid)
	return err
}

func (p *DB) GetFriendList(uid uint64) (friendUidList []uint64, err error) {
	rows, err := p.queryRows("SELECT friend_id FROM Friends WHERE user_id = ? AND status = 1", uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var friendID uint64
		err := rows.Scan(&friendID)
		if err != nil {
			return nil, err
		}
		friendUidList = append(friendUidList, friendID)
	}

	return friendUidList, nil
}

// 分配 seqid。数据库中会存储每个用户当前可分配的 seqid，从 1 开始，每次分配后 seqid + 1
func (p *DB) AllocateSeqId(uid uint64) (seqId uint64, err error) {
	row, err := p.queryRow("SELECT seq_id FROM SeqIds WHERE user_id = ?", uid)
	if err != nil {
		return 0, fmt.Errorf("queryRow: %w", err)
	}

	err = row.Scan(&seqId)
	if err != nil {
		if err != sql.ErrNoRows {
			return 0, fmt.Errorf("Scan: %w", err)
		}

		// 用户还没有分配过 seqId，需要初始化
		_, err = p.sqlExec( "INSERT INTO SeqIds (user_id, seq_id) VALUES (?, ?)", uid, 2)
		if err != nil {
			return 0, fmt.Errorf("sqlExec: %w", err)
		}
		return 1, nil
	}

	// 更新 seqId
	_, err = p.sqlExec("UPDATE SeqIds SET seq_id = ? WHERE user_id = ?", seqId + 1, uid)
	if err != nil {
		return 0, fmt.Errorf("sqlExec: %w", err)
	}

	return seqId, nil
}