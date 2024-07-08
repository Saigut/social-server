package user_mgmt

import (
	"errors"
	"fmt"
	"regexp"
	"social_server/src/app/common/types"
	"social_server/src/app/data"
	. "social_server/src/utils/log"
	"time"
	"unicode"
)

type UserMgmt struct {
	storage *data.DB
	cache *data.Cache
}

func NewUserMgmt(storage *data.DB, cache *data.Cache) *UserMgmt {
	return &UserMgmt{
		storage: storage,
		cache: cache,
	}
}

// 验证用户名格式
func validateUsername(username string) bool {
	if len(username) < 3 || len(username) > 20 {
		return false
	}
	for _, char := range username {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
			return false
		}
	}
	return true
}

// 验证密码格式
func (p *UserMgmt)ValidatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	return true

	//if len(password) < 8 || len(password) > 16 {
	//	return false
	//}
	//
	//var hasUpper, hasLower, hasNumber, hasSpecial bool
	//for _, char := range password {
	//	switch {
	//	case unicode.IsUpper(char):
	//		hasUpper = true
	//	case unicode.IsLower(char):
	//		hasLower = true
	//	case unicode.IsDigit(char):
	//		hasNumber = true
	//	case unicode.IsPunct(char) || unicode.IsSymbol(char):
	//		hasSpecial = true
	//	default:
	//		return false
	//	}
	//}
	//
	//typeCount := 0
	//if hasUpper || hasLower {
	//	typeCount++
	//}
	//if hasNumber {
	//	typeCount++
	//}
	//if hasSpecial {
	//	typeCount++
	//}
	//
	//return typeCount >= 2
}

// 验证电子邮件格式
func validateEmail(email string) bool {
	// 简单的电子邮件正则表达式
	emailRegex := `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

func (p *UserMgmt) UserInfoValidate(param *types.UmUserInfoValidateParam) (err error) {
	if !validateUsername(param.Username) {
		return errors.New("invalid username")
	}
	if !p.ValidatePassword(param.Passphase) {
		return errors.New("invalid password")
	}
	if !validateEmail(param.Email) {
		return errors.New("invalid email")
	}
	return nil
}

func (p *UserMgmt) IsUsernameExisted(userName string) (bool, error) {
	// 查询数据库
	isExist, err := p.storage.UserIsUsernameExisted(userName)
	if err != nil {
		return false, err
	}
	return isExist, nil
}

func (p *UserMgmt) UserAuthenticate(param *types.UmUserAuthenticateParam) (pass bool, err error) {
	// 查询缓存
	pass, err = p.cache.UserAuthenticate(param)
	if err != nil {
		_, ok := err.(*data.CacheNotFoundError)
		if !ok {
			return false, err
		}
	} else {
		return pass, nil
	}

	// 若无缓存，查询数据库，并更新缓存
	pass, err = p.storage.UserAuthenticate(param)
	if err != nil {
		return false, err
	}
	if pass {
		err = p.cache.CacheUserAuthenticate(&types.UmUserInfo{
			Uid:      0,
			Username: param.Username,
			Password: param.Passphase,
			Email:    "",
		})
		if err != nil {
			Log.Warn("cache user authenticate error: %v", err)
		}
	}

	return pass, nil
}

func (p *UserMgmt) UserGetInfoByUsername(username string) (userInfo *types.UmUserInfo, err error) {
	userInfo, err = p.storage.UserGetInfoByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("UserGetInfoByUsername: %v", err)
	}
	return userInfo, nil
}

func (p *UserMgmt) Register(param *types.UmRegisterParam) (err error) {
	_, err = p.storage.UserRegister(param)
	return err
}

func (p *UserMgmt) Unregister(param *types.UmUnregisterParam) (err error) {
	// 查询用户信息
	var userInfo *types.UmUserInfo
	userInfo, err = p.storage.UserGetInfo(param.Uid)
	if err != nil {
		return err
	}

	// 清除用户信息缓存
	err = p.cache.ClearCacheUserAuthenticate(userInfo.Username)
	if err != nil {
		Log.Error("clear cache user authenticate error: %v", err)
		return err
	}

	// 更新数据库
	err = p.storage.UserUnregister(param)
	if err != nil {
		return err
	}

	// 启动协程，5秒后再删一次缓存
	go func() {
		time.Sleep(5 * time.Second)
		err = p.cache.ClearCacheUserAuthenticate(userInfo.Username)
		if err != nil {
			Log.Error("clear cache user authenticate error: %v", err)
		}
	}()

	return nil
}

func (p *UserMgmt) UserUpdateInfo(uid uint64, nickname string, email string, avatar string, password string, newPassword string) (err error) {
	// 验证密码
	var userInfo *types.UmUserInfo
	if password != "" {
		userInfo, err = p.storage.UserGetInfo(uid)
		if err != nil {
			return fmt.Errorf("UserGetInfo: %w", err)
		}
		if userInfo.Password != password {
			return fmt.Errorf("wrong password")
		}
	}
	// 更新用户信息
	err = p.storage.UserUpdateInfo(uid, nickname, email, avatar, newPassword)
	if err != nil {
		return fmt.Errorf("UserUpdateInfo: %w", err)
	}
	if password != "" {
		err = p.cache.ClearCacheUserAuthenticate(userInfo.Username)
		if err != nil {
			Log.Error("clear cache user authenticate error: %v", err)
		}
	}
	return nil
}

func (p *UserMgmt) ContactGetList(uid uint64) (contactUidList []uint64, err error) {
	contactUidList, err = p.storage.ContactGetList(uid)
	if err != nil {
		return nil, err
	}
	return contactUidList, nil
}

func (p *UserMgmt) ContactGetRelation(uid uint64, contactUid uint64) (isMutualContact bool, remarkName string, err error) {
	isMutualContact, remarkName, err = p.storage.ContactGetRelation(uid, contactUid)
	if err != nil {
		return false, "", fmt.Errorf("ContactGetRelation: %w", err)
	}
	return isMutualContact, remarkName, nil
}

func (p *UserMgmt) ContactFind(username string) (userInfo *types.UmUserInfo, err error) {
	userInfo, err = p.storage.UserGetInfoByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("UserGetInfoByUsername: %v", err)
	}
	return userInfo, nil
}

func (p *UserMgmt) ContactGetInfo(uid uint64) (userInfo *types.UmUserInfo, err error) {
	userInfo, err = p.storage.UserGetInfo(uid)
	if err != nil {
		return nil, fmt.Errorf("UserGetInfoByUsername: %v", err)
	}
	return userInfo, nil
}

func (p *UserMgmt) ContactAccept(uid uint64, contactUid uint64) (err error) {
	// 更新数据库
	err = p.storage.ContactAccept(uid, contactUid)
	if err != nil {
		return fmt.Errorf("ContactAdd: %v", err)
	}
	return err
}

func (p *UserMgmt) ContactReject(uid uint64, contactUid uint64) (err error) {
	// 更新数据库
	err = p.storage.ContactReject(uid, contactUid)
	if err != nil {
		return fmt.Errorf("ContactReject: %v", err)
	}
	return err
}

func (p *UserMgmt) ContactDel(uid uint64, contactUid uint64) (err error) {
	// 更新数据库
	err = p.storage.ContactDel(uid, contactUid)
	if err != nil {
		return fmt.Errorf("ContactDel: %v", err)
	}
	return err
}

func (p *UserMgmt) GroupGetList(uid uint64) (ConvIdList []uint64, err error) {
	ConvIdList, err = p.storage.GroupGetList(uid)
	if err != nil {
		return nil, err
	}
	return ConvIdList, nil
}

func (p *UserMgmt) GroupGetInfo(groupId uint64) (groupInfo *types.UmGroupInfo, err error) {
	groupInfo, err = p.storage.GroupGetInfo(groupId)
	if err != nil {
		return nil, fmt.Errorf("GroupGetInfo: %v", err)
	}
	return groupInfo, nil
}

func (p *UserMgmt) GroupUpdateInfo(groupId uint64, groupName string, avatar string) (err error) {
	err = p.storage.GroupUpdateInfo(groupId, groupName, avatar)
	if err != nil {
		return fmt.Errorf("GroupUpdateInfo: %w", err)
	}
	return nil
}

func (p *UserMgmt) GroupFind(groupId uint64) (groupInfo *types.UmGroupInfo, err error) {
	groupInfo, err = p.storage.GroupGetInfo(groupId)
	if err != nil {
		return nil, fmt.Errorf("GroupGetInfo: %v", err)
	}
	return groupInfo, nil
}

func (p *UserMgmt) GroupCreate(uid uint64, groupName string) (ConvId uint64, err error) {
	// 更新数据库
	ConvId, err = p.storage.GroupCreate(uid, groupName)
	if err != nil {
		return 0, err
	}
	return ConvId, nil
}

func (p *UserMgmt) GroupDelete(uid uint64, groupId uint64) (err error) {
	// 更新数据库
	err = p.storage.GroupDelete(uid, groupId)
	if err != nil {
		return fmt.Errorf("GroupDelete: %w", err)
	}
	return nil
}

func (p *UserMgmt) GroupGetMemList(groupId uint64) (uidList []uint64, err error) {
	uidList, err = p.storage.GroupGetMemList(groupId)
	if err != nil {
		return nil, fmt.Errorf("GroupGetMemList: %w", err)
	}
	return uidList, nil
}

func (p *UserMgmt) GroupGetAdminList(groupId uint64) (uidList []uint64, err error) {
	uidList, err = p.storage.GroupGetAdminList(groupId)
	if err != nil {
		return nil, fmt.Errorf("GroupGetAdminList: %w", err)
	}
	return uidList, nil
}

func (p *UserMgmt) GroupIsOwner(groupId uint64, uid uint64) (isOwner bool, err error) {
	isOwner, err = p.storage.GroupIsOwner(groupId, uid)
	if err != nil {
		return false, fmt.Errorf("GroupIsOwner: %w", err)
	}
	return isOwner, nil
}

func (p *UserMgmt) GroupIsMem(groupId uint64, uid uint64) (inGroup bool, err error) {
	inGroup, err = p.storage.GroupIsMem(groupId, uid)
	if err != nil {
		return false, err
	}
	return inGroup, nil
}

func (p *UserMgmt) GroupLeave(groupId uint64, uid uint64) (err error) {
	// 更新数据库
	err = p.storage.GroupDelMem(groupId, uid)
	if err != nil {
		return fmt.Errorf("GroupDelMem: %w", err)
	}
	return nil
}

func (p *UserMgmt) GroupAddMem(groupId uint64, adminUid uint64, uid uint64) (err error) {
	// 检查 adminUid 是否为管理员
	isAdmin, err := p.storage.GroupIsAdmin(groupId, adminUid)
	if err != nil {
		return fmt.Errorf("GroupIsAdmin: %w", err)
	}
	if !isAdmin {
		return fmt.Errorf("uid is not the admin of the group")
	}

	// 更新数据库
	err = p.storage.GroupAddMem(groupId, uid, 0)
	if err != nil {
		return fmt.Errorf("GroupAddMem: %w", err)
	}

	return nil
}

func (p *UserMgmt) GroupDelMem(groupId uint64, adminUid uint64, uid uint64) (err error) {
	// 检查 adminUid 是否为管理员
	isAdmin, err := p.storage.GroupIsAdmin(groupId, adminUid)
	if err != nil {
		return fmt.Errorf("GroupIsAdmin: %w", err)
	}
	if !isAdmin {
		return fmt.Errorf("uid is not the admin of the group")
	}

	// 更新数据库
	err = p.storage.GroupDelMem(groupId, uid)
	if err != nil {
		return fmt.Errorf("GroupDelMem: %w", err)
	}

	return nil
}

func (p *UserMgmt) GroupAccept(groupId uint64, adminUid uint64, uid uint64) (err error) {
	// 检查 adminUid 是否为管理员
	isAdmin, err := p.storage.GroupIsAdmin(groupId, adminUid)
	if err != nil {
		return fmt.Errorf("GroupIsAdmin: %w", err)
	}
	if !isAdmin {
		// GroupIgnore
		err = p.storage.GroupIgnore(groupId, uid)
		Log.Warn("uid is not the admin of the group")
		return nil
	}
	// 更新数据库
	err = p.storage.GroupAccept(groupId, uid)
	if err != nil {
		return fmt.Errorf("GroupAccept: %w", err)
	}
	return nil
}

func (p *UserMgmt) GroupReject(groupId uint64, adminUid uint64, uid uint64) (err error) {
	// 检查 adminUid 是否为管理员
	isAdmin, err := p.storage.GroupIsAdmin(groupId, adminUid)
	if err != nil {
		return fmt.Errorf("GroupIsAdmin: %w", err)
	}
	if !isAdmin {
		// GroupIgnore
		err = p.storage.GroupIgnore(groupId, uid)
		Log.Warn("uid is not the admin of the group")
		return nil
	}
	// 更新数据库
	err = p.storage.GroupReject(groupId, uid)
	if err != nil {
		return fmt.Errorf("GroupReject: %w", err)
	}
	return nil
}

func (p *UserMgmt) GroupUpdateMem(groupId uint64, adminUid uint64, uid uint64, role uint) (err error) {
	// 检查 adminUid 是否为管理员
	isAdmin, err := p.storage.GroupIsAdmin(groupId, adminUid)
	if err != nil {
		return fmt.Errorf("GroupIsAdmin: %w", err)
	}
	if !isAdmin {
		return fmt.Errorf("uid is not the admin of the group")
	}

	// 更新数据库
	err = p.storage.GroupUpdateMem(groupId, uid, role)
	if err != nil {
		return fmt.Errorf("GroupUpdateMem: %w", err)
	}

	return nil
}
