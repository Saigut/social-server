package data

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"social_server/src/app/common/types"
	"social_server/src/gen/grpc"
	. "social_server/src/utils/log"
	"strings"
	"sync"
	"time"
)

// tb_users table structure
type User struct {
	UserID    uint64
	Password  string
	Username  string
	Nickname  string
	Email     string
	Avatar    string
	CreatedAt time.Time
}

// tb_user_contacts table structure
type Contact struct {
	UserID    uint64
	ContactID uint64
	IsMutualContact bool
	CreatedAt time.Time
}

// tb_groups table structure
type ChatGroup struct {
	GroupID   uint64
	GroupName string
	OwnerUid  uint64
	Avatar    string
	MemCount   uint64
	CreatedAt time.Time
}

// tb_group_members table structure
type ChatGroupMember struct {
	GroupID  uint64
	UserID   uint64
	Role     uint
	JoinedAt time.Time
}

// Messages table structure
type Message struct {
	MessageID   uint64
	SenderID    uint64
	ReceiverID  uint64
	GroupID     uint64
	Content     string
	MessageType int
	SentAt      time.Time
}

type InboxMsg struct {
	UserID      uint64
	SeqID     uint64
	ConvMsgId uint64
	RandMsgId uint64
	SenderID    uint64
	ReceiverID  sql.NullInt64
	GroupID     sql.NullInt64
	MessageType int
	Content     string
	ReadMsgId uint64
	SentAt      time.Time
	IsRead    bool
	Status     uint32
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

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=UTC", dbUser, dbPassword, dbHost, dbPort, dbName)

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

func (p *DB) sqlTxExec(tx *sql.Tx, sqlStr string, args ...interface{}) (sql.Result, error) {
	ret, err := p.withReconnectHandling(func() (interface{}, error) {
		return tx.Exec(sqlStr, args...)
	})
	if err != nil {
		return nil, fmt.Errorf("withReconnectHandling: %w", err)
	}
	return ret.(sql.Result), nil
}

func (p *DB) sqlTxCommit(tx *sql.Tx) error {
	_, err := p.withReconnectHandling(func() (interface{}, error) {
		return nil, tx.Commit()
	})
	if err != nil {
		return fmt.Errorf("withReconnectHandling: %w", err)
	}
	return nil
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
	msg.SenderUid = rowMsg.SenderID
	msg.SentTsMs = uint64(rowMsg.SentAt.UnixNano() / 1e6)
	msg.MsgContent = rowMsg.Content
	msg.MsgType = gen_grpc.ChatMsgType(rowMsg.MessageType)
	msg.ReadMsgId = rowMsg.ReadMsgId
	return msg, nil
}

func convertDbMsgToChatMsgOfConv(rowMsg InboxMsg) (msg types.ChatMsgOfConv, err error) {
	msg.Msg, err = convertDbMsgToChatMsg(rowMsg)
	if err != nil {
		return msg, fmt.Errorf("convertDbMsgToChatMsg: %w", err)
	}

	msg.SeqId = rowMsg.SeqID
	msg.ConvMsgId = rowMsg.ConvMsgId
	msg.RandMsgId = rowMsg.RandMsgId
	msg.IsRead = rowMsg.IsRead
	msg.Status = rowMsg.Status

	if rowMsg.GroupID.Valid {
		msg.ReceiverId.PeerIdType = types.EmPeerIdType_GroupId
		msg.ReceiverId.GroupId = uint64(rowMsg.GroupID.Int64)
	} else if rowMsg.ReceiverID.Valid {
		msg.ReceiverId.PeerIdType = types.EmPeerIdType_Uid
		msg.ReceiverId.Uid = uint64(rowMsg.ReceiverID.Int64)
	} else {
		return msg, errors.New("invalid peer id")
	}

	return msg, nil
}


func (p *DB) AllocateSeqId(uid uint64) (seqId uint64, err error) {
	row, err := p.queryRow("SELECT seq_id FROM tb_seq_id_user WHERE user_id = ?", uid)
	if err != nil {
		return 0, fmt.Errorf("queryRow: %w", err)
	}

	err = row.Scan(&seqId)
	if err != nil {
		if err != sql.ErrNoRows {
			return 0, fmt.Errorf("Scan: %w", err)
		}

		// 还没有分配过 seqId，需要初始化
		_, err = p.sqlExec( "INSERT INTO tb_seq_id_user (user_id, seq_id) VALUES (?, ?)", uid, 2)
		if err != nil {
			return 0, fmt.Errorf("sqlExec: %w", err)
		}
		return 1, nil
	}

	// 更新 seqId
	_, err = p.sqlExec("UPDATE tb_seq_id_user SET seq_id = ? WHERE user_id = ?", seqId + 1, uid)
	if err != nil {
		return 0, fmt.Errorf("sqlExec: %w", err)
	}

	return seqId, nil
}

func (p *DB) AllocateChatSeqId(uid1 uint64, uid2 uint64) (seqId uint64, err error) {
	if uid1 > uid2 {
		uid1, uid2 = uid2, uid1
	}
	row, err := p.queryRow("SELECT seq_id FROM tb_seq_id_chat WHERE user1_id = ? AND user2_id = ?", uid1, uid2)
	if err != nil {
		return 0, fmt.Errorf("queryRow: %w", err)
	}

	err = row.Scan(&seqId)
	if err != nil {
		if err != sql.ErrNoRows {
			return 0, fmt.Errorf("Scan: %w", err)
		}

		// 还没有分配过 seqId，需要初始化
		_, err = p.sqlExec( "INSERT INTO tb_seq_id_chat (user1_id, user2_id, seq_id) VALUES (?, ?, 2)", uid1, uid2)
		if err != nil {
			return 0, fmt.Errorf("sqlExec: %w", err)
		}
		return 1, nil
	}

	// 更新 seqId
	_, err = p.sqlExec("UPDATE tb_seq_id_chat SET seq_id = ? WHERE user1_id = ? AND user2_id = ?", seqId + 1, uid1, uid2)
	if err != nil {
		return 0, fmt.Errorf("sqlExec: %w", err)
	}

	return seqId, nil
}

func (p *DB) AllocateGroupSeqId(groupId uint64) (seqId uint64, err error) {
	row, err := p.queryRow("SELECT seq_id FROM tb_seq_id_group WHERE group_id = ?", groupId)
	if err != nil {
		return 0, fmt.Errorf("queryRow: %w", err)
	}

	err = row.Scan(&seqId)
	if err != nil {
		if err != sql.ErrNoRows {
			return 0, fmt.Errorf("Scan: %w", err)
		}

		// 还没有分配过 seqId，需要初始化
		_, err = p.sqlExec( "INSERT INTO tb_seq_id_group (group_id, seq_id) VALUES (?, ?)", groupId, 2)
		if err != nil {
			return 0, fmt.Errorf("sqlExec: %w", err)
		}
		return 1, nil
	}

	// 更新 seqId
	_, err = p.sqlExec("UPDATE tb_seq_id_group SET seq_id = ? WHERE group_id = ?", seqId + 1, groupId)
	if err != nil {
		return 0, fmt.Errorf("sqlExec: %w", err)
	}

	return seqId, nil
}

// UserMgmt
func (p *DB) UserIsUsernameExisted(username string) (bool, error) {
	row, err := p.queryRow("SELECT COUNT(*) FROM tb_users WHERE LOWER(username) = LOWER(?)", username)
	var count int
	err = row.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (p *DB) UserAuthenticate(param *types.UmUserAuthenticateParam) (bool, error) {
	row, err := p.queryRow("SELECT COUNT(*) FROM tb_users WHERE username = ? AND password = ?", param.Username, param.Passphase)
	var count int
	err = row.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (p *DB) UserRegister(param *types.UmRegisterParam) (uint64, error) {
	res, err := p.sqlExec("INSERT INTO tb_users (password, username, nickname, email, avatar) VALUES (?, ?, ?, ?, ?)",
		param.Passwd, param.Username, param.Nickname, param.Email, param.Avatar)
	if err != nil {
		return 0, fmt.Errorf("sqlExec: %w", err)
	}
	// 获取新增用户的id
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("LastInsertId: %w", err)
	}
	return uint64(id), nil
}

func (p *DB) UserUnregister(param *types.UmUnregisterParam) (error) {
	_, err := p.sqlExec("DELETE FROM tb_users WHERE user_id = ?", param.Uid)
	return err
}

func (p *DB) UserGetInfo(uid uint64) (user *types.UmUserInfo, err error) {
	row, err := p.queryRow("SELECT user_id, password, username, nickname, email, avatar FROM tb_users WHERE user_id = ?", uid)
	if err != nil {
		return nil, err
	}
	var userRow User
	err = row.Scan(&userRow.UserID, &userRow.Password, &userRow.Username, &userRow.Nickname, &userRow.Email, &userRow.Avatar)
	if err != nil {
		return nil, err
	}
	return &types.UmUserInfo{
		Uid:      userRow.UserID,
		Password: userRow.Password,
		Username: userRow.Username,
		Nickname: userRow.Nickname,
		Email:    userRow.Email,
		Avatar:   userRow.Avatar,
	}, nil
}

func (p *DB) UserGetInfoByUsername(username string) (user *types.UmUserInfo, err error) {
	row, err := p.queryRow("SELECT user_id, password, username, nickname, email, avatar FROM tb_users WHERE LOWER(username) = LOWER(?)", username)
	if err != nil {
		return nil, err
	}
	var userRow User
	err = row.Scan(&userRow.UserID, &userRow.Password, &userRow.Username, &userRow.Nickname, &userRow.Email, &userRow.Avatar)
	if err != nil {
		return nil, err
	}
	return &types.UmUserInfo{
		Uid:      userRow.UserID,
		Password: userRow.Password,
		Username: userRow.Username,
		Nickname: userRow.Nickname,
		Email:    userRow.Email,
		Avatar:   userRow.Avatar,
	}, nil
}

func (p *DB) UserUpdateInfo(uid uint64, nickname string, email string, avatar string, password string) (err error) {
	var fields []string
	var args []interface{}

	if nickname != "" {
		fields = append(fields, "nickname = ?")
		args = append(args, nickname)
	}

	if email != "" {
		fields = append(fields, "email = ?")
		args = append(args, email)
	}

	if avatar != "" {
		fields = append(fields, "avatar = ?")
		args = append(args, avatar)
	}

	if password != "" {
		fields = append(fields, "password = ?")
		args = append(args, password)
	}

	// 如果没有字段需要更新，直接返回 nil
	if len(fields) == 0 {
		return nil
	}

	// 构建完整的 SQL 语句
	query := fmt.Sprintf("UPDATE tb_users SET %s WHERE user_id = ?", strings.Join(fields, ", "))
	args = append(args, uid)

	// 执行 SQL 语句
	_, err = p.sqlExec(query, args...)
	if err != nil {
		return fmt.Errorf("sqlExec: %w", err)
	}

	return nil
}

func (p *DB) ContactGetList(uid uint64) (contactUidList []uint64, err error) {
	rows, err := p.queryRows("SELECT contact_id FROM tb_user_contacts WHERE user_id = ?", uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var contactID uint64
		err := rows.Scan(&contactID)
		if err != nil {
			return nil, err
		}
		contactUidList = append(contactUidList, contactID)
	}

	return contactUidList, nil
}

func (p *DB) ContactGetRelation(uid uint64, contactUid uint64) (isMutualContact bool, remarkName string, err error) {
	row, err := p.queryRow("SELECT is_mutual_contact, remark_name FROM tb_user_contacts WHERE user_id = ? AND contact_id = ?", uid, contactUid)
	if err != nil {
		return false, "", fmt.Errorf("queryRow: %w", err)
	}

	err = row.Scan(&isMutualContact, &remarkName)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, "", nil
		}
		return false, "", fmt.Errorf("Scan: %w", err)
	}
	return isMutualContact, remarkName, nil
}

func (p *DB) ContactAdd(uid uint64, contactUid uint64) error {
	tx, err := p.db.Begin() // 开启事务
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // 重新 panic
		} else if err != nil {
			tx.Rollback() // 有错误回滚
		}
	}()

	// 插入或更新记录
	_, err = p.sqlTxExec(tx, `
		INSERT INTO tb_user_contacts (user_id, contact_id, is_mutual_contact)
		VALUES (?, ?, true)
		ON DUPLICATE KEY UPDATE is_mutual_contact = VALUES(is_mutual_contact)
	`, uid, contactUid)
	if err != nil {
		return fmt.Errorf("sqlTxExec: %w", err)
	}

	// 插入或更新反向关系
	_, err = p.sqlTxExec(tx, `
		INSERT INTO tb_user_contacts (user_id, contact_id, is_mutual_contact)
		VALUES (?, ?, true)
		ON DUPLICATE KEY UPDATE is_mutual_contact = VALUES(is_mutual_contact)
	`, contactUid, uid)
	if err != nil {
		return fmt.Errorf("sqlTxExec: %w", err)
	}

	// 提交事务
	err = p.sqlTxCommit(tx)
	if err != nil {
		return fmt.Errorf("sqlTxCommit: %w", err)
	}

	return nil
}

func (p *DB) ContactAccept(uid uint64, contactUid uint64) (error) {
	err := p.ContactAdd(uid, contactUid)
	if err != nil {
		return fmt.Errorf("ContactAdd: %w", err)
	}
	// 在 tb_user_inbox 表中更新好友请求消息状态
	_, err = p.sqlExec("UPDATE tb_user_inbox SET status = 1 WHERE user_id = ? AND sender_id = ? AND receiver_id = ? AND status = 0",
		uid, contactUid, uid)
	if err != nil {
		return fmt.Errorf("sqlExec: %w", err)
	}
	return nil
}

func (p *DB) ContactDel(uid uint64, contactUid uint64) (error) {
	_, err := p.sqlExec("DELETE FROM tb_user_contacts WHERE user_id = ? AND contact_id = ?", uid, contactUid)
	if err != nil {
		return fmt.Errorf("sqlExec: %w", err)
	}
	_, err = p.sqlExec("DELETE FROM tb_user_contacts WHERE user_id = ? AND contact_id = ?", contactUid, uid)
	if err != nil {
		return fmt.Errorf("sqlExec: %w", err)
	}
	_, err = p.sqlExec("DELETE FROM tb_user_inbox WHERE user_id = ? AND sender_id = ? AND receiver_id = ?",
		uid, contactUid, uid)
	if err != nil {
		return fmt.Errorf("sqlExec: %w", err)
	}
	_, err = p.sqlExec("DELETE FROM tb_user_inbox WHERE user_id = ? AND sender_id = ? AND receiver_id = ?",
		uid, uid, contactUid)
	if err != nil {
		return fmt.Errorf("sqlExec: %w", err)
	}
	return nil
}

func (p *DB) ContactReject(uid uint64, contactUid uint64) (error) {
	_, err := p.sqlExec("UPDATE tb_user_inbox SET status = 2 WHERE user_id = ? AND sender_id = ?  AND receiver_id = ? AND status = 0",
		uid, contactUid, uid)
	if err != nil {
		return fmt.Errorf("sqlExec: %w", err)
	}
	return nil
}

func (p *DB) GroupGetList(uid uint64) (ConvIdList []uint64, err error) {
	rows, err := p.queryRows("SELECT group_id FROM tb_group_members WHERE user_id = ?", uid)
	if err != nil {
		return nil, fmt.Errorf("queryRows: %w", err)
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

func (p *DB) GroupGetInfo(groupId uint64) (groupInfo *types.UmGroupInfo, err error) {
	row, err := p.queryRow("SELECT group_id, group_name, owner_id, avatar, mem_count, created_at FROM tb_groups WHERE group_id = ?", groupId)
	if err != nil {
		return nil, fmt.Errorf("queryRow: %w", err)
	}
	var group ChatGroup
	err = row.Scan(&group.GroupID, &group.GroupName, &group.OwnerUid, &group.Avatar, &group.MemCount, &group.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("Scan: %w", err)
	}
	return &types.UmGroupInfo{
		GroupId:    group.GroupID,
		GroupName:  group.GroupName,
		OwnerUid:   group.OwnerUid,
		Avatar:     group.Avatar,
		MemCount:   group.MemCount,
		CreateTsMs: uint64(group.CreatedAt.UnixNano() / 1e6),
	}, nil
}

func (p *DB) GroupUpdateInfo(groupId uint64, groupName string, avatar string) (err error) {
	// 动态构建 SQL 语句和参数列表
	var fields []string
	var args []interface{}

	if groupName != "" {
		fields = append(fields, "group_name = ?")
		args = append(args, groupName)
	}

	if avatar != "" {
		fields = append(fields, "avatar = ?")
		args = append(args, avatar)
	}

	// 如果没有字段需要更新，直接返回 nil
	if len(fields) == 0 {
		return nil
	}

	// 构建完整的 SQL 语句
	query := fmt.Sprintf("UPDATE tb_groups SET %s WHERE group_id = ?", strings.Join(fields, ", "))
	args = append(args, groupId)

	// 执行 SQL 语句
	_, err = p.sqlExec(query, args...)
	if err != nil {
		return fmt.Errorf("sqlExec: %w", err)
	}

	return nil
}


func (p *DB) GroupCreate(uid uint64, groupName string) (ConvId uint64, err error) {
	res, err := p.sqlExec("INSERT INTO tb_groups (group_name, owner_id, mem_count) VALUES (?, ?, 1)", groupName, uid)
	if err != nil {
		return 0, err
	}

	groupID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	_, err = p.sqlExec("INSERT INTO tb_group_members (group_id, user_id, role) VALUES (?, ?, 1)", groupID, uid)
	return uint64(groupID), err
}

func (p *DB) GroupDelete(uid uint64, groupId uint64) (err error) {
	isOwner, err := p.GroupIsOwner(groupId, uid)
	if err != nil {
		return fmt.Errorf("GroupIsOwner: %w", err)
	}
	if !isOwner {
		return fmt.Errorf("uid is not the owner of the group")
	}
	// 清空群成员
	_, err = p.sqlExec("DELETE FROM tb_group_members WHERE group_id = ?", groupId)
	if err != nil {
		return fmt.Errorf("sqlExec: %w", err)
	}
	// 删除群聊
	_, err = p.sqlExec("DELETE FROM tb_groups WHERE group_id = ?", groupId)
	if err != nil {
		return fmt.Errorf("sqlExec: %w", err)
	}
	return nil
}

func (p *DB) GroupGetMemList(groupId uint64) (memUidList []uint64, err error) {
	rows, err := p.queryRows("SELECT user_id FROM tb_group_members WHERE group_id = ?", groupId)
	if err != nil {
		return nil, fmt.Errorf("queryRows: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var uid uint64
		err := rows.Scan(&uid)
		if err != nil {
			return nil, err
		}
		memUidList = append(memUidList, uid)
	}
	return memUidList, nil
}

func (p *DB) GroupGetAdminList(groupId uint64) (adminUid []uint64, err error) {
	rows, err := p.queryRows("SELECT user_id FROM tb_group_members WHERE group_id = ? AND (role = 1 OR role = 2)", groupId)
	if err != nil {
		return nil, fmt.Errorf("queryRows: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var uid uint64
		err := rows.Scan(&uid)
		if err != nil {
			return nil, fmt.Errorf("Scan: %w", err)
		}
		adminUid = append(adminUid, uid)
	}

	return adminUid, nil
}

func (p *DB) GroupIsOwner(groupId uint64, uid uint64) (isOwner bool, err error) {
	row, err := p.queryRow("SELECT COUNT(*) FROM tb_group_members WHERE group_id = ? AND user_id = ? AND role = 1", groupId, uid)
	var count int
	err = row.Scan(&count)
	if err != nil {
		return false, fmt.Errorf("queryRows: %w", err)
	}
	return count > 0,nil
}

func (p *DB) GroupIsAdmin(groupId uint64, uid uint64) (isAdmin bool, err error) {
	row, err := p.queryRow("SELECT COUNT(*) FROM tb_group_members WHERE group_id = ? AND user_id = ? AND (role = 1 OR role = 2)", groupId, uid)
	var count int
	err = row.Scan(&count)
	if err != nil {
		return false, fmt.Errorf("queryRows: %w", err)
	}
	return count > 0, nil
}

func (p *DB) GroupIsMem(groupId uint64, uid uint64) (inGroup bool, err error) {
	row, err := p.queryRow("SELECT COUNT(*) FROM tb_group_members WHERE group_id = ? AND user_id = ?", groupId, uid)
	var count int
	err = row.Scan(&count)
	if err != nil {
		return false, fmt.Errorf("queryRows: %w", err)
	}
	return count > 0, nil
}

func (p *DB) GroupClearMsg(groupId uint64, uid uint64) (err error) {
	// 删除群聊消息
	_, err = p.sqlExec("DELETE FROM tb_user_inbox WHERE user_id = ? AND group_id = ?", uid, groupId)
	if err != nil {
		return fmt.Errorf("sqlExec: %w", err)
	}
	return nil
}

func (p *DB) GroupLeave(groupId uint64, uid uint64) (err error) {
	// 删除群成员
	_, err = p.sqlExec("DELETE FROM tb_group_members WHERE group_id = ? AND user_id = ?", groupId, uid)
	if err != nil {
		return fmt.Errorf("Member sqlExec: %w", err)
	}
	// 更新群成员数量
	_, err = p.sqlExec("UPDATE tb_groups SET mem_count = mem_count - 1 WHERE group_id = ?", groupId)
	if err != nil {
		return fmt.Errorf("Count sqlExec: %w", err)
	}
	// 为此用户删除群聊消息
	_, err = p.sqlExec("DELETE FROM tb_user_inbox WHERE user_id = ? AND group_id = ?", uid, groupId)
	if err != nil {
		return fmt.Errorf("Inbox sqlExec: %w", err)
	}
	return nil
}

func (p *DB) GroupAddMem(groupId uint64, uid uint64, role uint) (err error) {
	// 添加群成员
	_, err = p.sqlExec("INSERT INTO tb_group_members (group_id, user_id, role) VALUES (?, ?, ?)", groupId, uid, role)
	if err != nil {
		return fmt.Errorf("sqlExec: %w", err)
	}
	// 更新群成员数量
	_, err = p.sqlExec("UPDATE tb_groups SET mem_count = mem_count + 1 WHERE group_id = ?", groupId)
	if err != nil {
		return fmt.Errorf("Count sqlExec: %w", err)
	}
	return nil
}

func (p *DB) GroupDelMem(groupId uint64, uid uint64) (err error) {
	// 删除群成员
	_, err = p.sqlExec("DELETE FROM tb_group_members WHERE group_id = ? AND user_id = ?", groupId, uid)
	if err != nil {
		return fmt.Errorf("sqlExec: %w", err)
	}
	if _, err = p.sqlExec("UPDATE tb_groups SET mem_count = mem_count - 1 WHERE group_id = ?", groupId); err != nil {
		return fmt.Errorf("Count sqlExec: %w", err)
	}
	return nil
}

func (p *DB) GroupAccept(groupId uint64, uid uint64) (err error) {
	err = p.GroupAddMem(groupId, uid, 0)
	if err != nil {
		return fmt.Errorf("GroupAddMem: %w", err)
	}
	// 更新入群请求状态
	_, err = p.sqlExec("UPDATE tb_user_inbox SET status = 1 WHERE group_id = ? AND sender_id = ?", groupId, uid)
	if err != nil {
		return fmt.Errorf("sqlExec: %w", err)
	}
	return nil
}

func (p *DB) GroupReject(groupId uint64, uid uint64) (err error) {
	// 在 tb_user_inbox 表中更新入群请求消息状态
	_, err = p.sqlExec("UPDATE tb_user_inbox SET status = 2 WHERE group_id = ? AND sender_id = ?", groupId, uid)
	if err != nil {
		return fmt.Errorf("sqlExec: %w", err)
	}
	return nil
}

func (p *DB) GroupIgnore(groupId uint64, uid uint64) (err error) {
	// 更新入群请求状态
	_, err = p.sqlExec("UPDATE tb_user_inbox SET status = 3 WHERE group_id = ? AND sender_id = ?", groupId, uid)
	if err != nil {
		return fmt.Errorf("sqlExec: %w", err)
	}
	return nil
}

func (p *DB) GroupUpdateMem(groupId uint64, uid uint64, role uint) (err error) {
	// 更新群成员
	_, err = p.sqlExec("UPDATE tb_group_members SET role = ? WHERE group_id = ? AND user_id = ?", role, groupId, uid)
	if err != nil {
		return fmt.Errorf("sqlExec: %w", err)
	}
	return nil
}

func (p *DB) ChatSendMsgToUser(uid uint64, convMsg types.ChatMsgOfConv) (err error) {
	var receiverId interface{}
	var groupId interface{}
	if convMsg.ReceiverId.PeerIdType == types.EmPeerIdType_Uid {
		receiverId = convMsg.ReceiverId.Uid
		groupId = nil
		if convMsg.Msg.MsgType == gen_grpc.ChatMsgType_emChatMsgType_MarkRead {
			// 删除 sender 发向 receiver 的已读消息
			_, err = p.sqlExec("DELETE FROM tb_user_inbox WHERE user_id = ? AND sender_id = ?  AND receiver_id = ? AND message_type = ? AND read_msg_id <= ?",
				uid, convMsg.Msg.SenderUid, receiverId, gen_grpc.ChatMsgType_emChatMsgType_MarkRead, convMsg.Msg.ReadMsgId)
			if err != nil {
				Log.Warn("sqlExec: %s", err)
			}
		}
	} else {
		groupId = convMsg.ReceiverId.GroupId
		receiverId = nil
		if convMsg.Msg.MsgType == gen_grpc.ChatMsgType_emChatMsgType_MarkRead {
			// 删除 sender 发向 group 的已读消息
			_, err = p.sqlExec("DELETE FROM tb_user_inbox WHERE user_id = ? AND sender_id = ?  AND group_id = ? AND message_type = ? AND read_msg_id <= ?",
				uid, convMsg.Msg.SenderUid, groupId, gen_grpc.ChatMsgType_emChatMsgType_MarkRead, convMsg.Msg.ReadMsgId)
			if err != nil {
				Log.Warn("sqlExec: %s", err)
			}
		}
	}

	// 分配 seqId
	var seqId uint64
	seqId, err = p.AllocateSeqId(uid)
	if err != nil {
		return fmt.Errorf("AllocateSeqId: %w", err)
	}

	// 添加消息
	_, err = p.sqlExec("INSERT INTO tb_user_inbox (user_id, seq_id, conv_msg_id, rand_msg_id, sender_id, receiver_id, group_id, content, message_type, read_msg_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		uid, seqId, convMsg.ConvMsgId, convMsg.RandMsgId, convMsg.Msg.SenderUid, receiverId, groupId, convMsg.Msg.MsgContent, convMsg.Msg.MsgType, convMsg.Msg.ReadMsgId)
	if err != nil {
		return fmt.Errorf("user sqlExec: %w", err)
	}

	return err
}

func (p *DB) ChatSendMsg(convMsg types.ChatMsgOfConv) (err error) {
	if convMsg.ReceiverId.PeerIdType == types.EmPeerIdType_Uid {
		// 分配 msgId
		convMsg.ConvMsgId, err = p.AllocateChatSeqId(convMsg.Msg.SenderUid, convMsg.ReceiverId.Uid)
		if err != nil {
			return fmt.Errorf("AllocateChatSeqId: %w", err)
		}

		// 添加消息
		err = p.ChatSendMsgToUser(convMsg.ReceiverId.Uid, convMsg)
		if err != nil {
			return fmt.Errorf("Receiver ChatSendMsgTo: %w", err)
		}

		// 也为发送者添加消息
		err = p.ChatSendMsgToUser(convMsg.Msg.SenderUid, convMsg)
		if err != nil {
			return fmt.Errorf("Sender ChatSendMsgTo: %w", err)
		}

	} else {
		// 获取群员群员列表
		memberList, err := p.GroupGetMemList(convMsg.ReceiverId.GroupId)
		if err != nil {
			return fmt.Errorf("GroupGetMemList: %w", err)
		}

		// 分配 msgId
		convMsg.ConvMsgId, err = p.AllocateGroupSeqId(convMsg.ReceiverId.GroupId)
		if err != nil {
			return fmt.Errorf("AllocateGroupSeqId: %w", err)
		}

		for _, memberUid := range memberList {
			// 添加消息
			err = p.ChatSendMsgToUser(memberUid, convMsg)
			if err != nil {
				return fmt.Errorf("Group ChatSendMsgTo: %w", err)
			}
		}
	}

	return err
}

func (p *DB) ChatSendMsgToAdmins(convMsg types.ChatMsgOfConv) (err error) {

	if convMsg.ReceiverId.PeerIdType != types.EmPeerIdType_GroupId {
		return fmt.Errorf("peer id type must be group id")
	}

	// 获取管理员列表
	adminUidList, err := p.GroupGetAdminList(convMsg.ReceiverId.GroupId)
	if err != nil {
		return fmt.Errorf("GroupGetAdminList: %w", err)
	}

	// 分配 msgId
	convMsg.ConvMsgId, err = p.AllocateGroupSeqId(convMsg.ReceiverId.GroupId)
	if err != nil {
		return fmt.Errorf("AllocateGroupSeqId: %w", err)
	}

	for _, adminUid := range adminUidList {
		// 分配 seqId
		seqId, err := p.AllocateSeqId(adminUid)
		if err != nil {
			return fmt.Errorf("AllocateSeqId: %w", err)
		}

		// 添加消息
		_, err = p.sqlExec( "INSERT INTO tb_user_inbox (user_id, seq_id, conv_msg_id, rand_msg_id, sender_id, receiver_id, group_id, content, message_type, read_msg_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			adminUid, seqId, convMsg.ConvMsgId, convMsg.RandMsgId, convMsg.Msg.SenderUid, nil, convMsg.ReceiverId.GroupId, convMsg.Msg.MsgContent, convMsg.Msg.MsgType, convMsg.Msg.ReadMsgId)
		if err != nil {
			return fmt.Errorf("sqlExec: %w", err)
		}
	}

	return nil
}

func (p *DB) ChatMarkRead(uid uint64, contactId uint64, readMsgId uint64) (err error) {
	_, err = p.sqlExec("UPDATE tb_user_inbox SET is_read = 1 WHERE user_id = ? AND (sender_id = ? AND receiver_id = ?) AND conv_msg_id <= ? AND is_read = 0",
		uid, contactId, uid, readMsgId)
	if err != nil {
		return fmt.Errorf("sqlExec: %w", err)
	}
	_, err = p.sqlExec("UPDATE tb_user_inbox SET is_read = 1 WHERE user_id = ? AND (sender_id = ? AND receiver_id = ?) AND conv_msg_id <= ? AND is_read = 0",
		contactId, contactId, uid, readMsgId)
	if err != nil {
		return fmt.Errorf("sqlExec: %w", err)
	}
	return nil
}

func (p *DB) ChatReadGroupMsg(uid uint64, groupId uint64, readMsgId uint64) (err error) {
	_, err = p.sqlExec("UPDATE tb_user_inbox SET is_read = 1 WHERE user_id = ? AND group_id = ? AND conv_msg_id <= ? AND is_read = 0", uid, groupId, readMsgId)
	if err != nil {
		return fmt.Errorf("sqlExec: %w", err)
	}
	return nil
}

func (p *DB) ChatGetMsgList(uid uint64, seqId uint64) (msgs []types.ChatMsgOfConv, err error) {
	// 查询。按 seqId 升序排列
	rows, err := p.queryRows(`
		SELECT user_id, seq_id, conv_msg_id, rand_msg_id, sender_id, receiver_id, group_id, content, message_type, read_msg_id, sent_at, is_read, status
		FROM tb_user_inbox WHERE user_id = ? AND seq_id > ? ORDER BY seq_id ASC`,
		uid, seqId,
	)
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
		err = rows.Scan(
			&rowMsg.UserID,
			&rowMsg.SeqID,
			&rowMsg.ConvMsgId,
			&rowMsg.RandMsgId,
			&rowMsg.SenderID,
			&rowMsg.ReceiverID,
			&rowMsg.GroupID,
			&rowMsg.Content,
			&rowMsg.MessageType,
			&rowMsg.ReadMsgId,
			&rowMsg.SentAt,
			&rowMsg.IsRead,
			&rowMsg.Status,
		)
		msg, err = convertDbMsgToChatMsgOfConv(rowMsg)

		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}

	return msgs, nil
}
