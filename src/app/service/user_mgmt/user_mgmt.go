package user_mgmt

import (
	"errors"
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
func validatePassword(password string) bool {
	if len(password) < 8 || len(password) > 32 {
		return false
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
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
	if !validatePassword(param.Passphase) {
		return errors.New("invalid password")
	}
	if !validateEmail(param.Email) {
		return errors.New("invalid email")
	}
	return nil
}

func (p *UserMgmt) IsUsernameExisted(userName string) (bool, error) {
	// 查询数据库
	isExist, err := p.storage.IsUsernameExisted(userName)
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
			Uid:       0,
			Username:  param.Username,
			Passphase: param.Passphase,
			Email:     "",
		})
		if err != nil {
			Log.Debug("cache user authenticate error: %v", err)
		}
	}

	return pass, nil
}

func (p *UserMgmt) Register(param *types.UmRegisterParam) (err error) {
	err = p.storage.Register(param)
	return err
}

func (p *UserMgmt) Unregister(param *types.UmUnregisterParam) (err error) {
	// 查询用户信息
	var userInfo *types.UmUserInfo
	userInfo, err = p.storage.GetUserInfo(param.Uid)
	if err != nil {
		return err
	}
	// 清除用户信息缓存
	err = p.cache.ClearCacheUserAuthenticate(userInfo.Username)
	if err != nil {
		Log.Error("clear cache user authenticate error: %v", err)
		return err
	}
	// 修改数据库
	err = p.storage.Unregister(param)
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

func (p *UserMgmt) AddFriends(param *types.UmAddFriendsParam) (err error) {
	err = p.storage.AddFriends(param)
	return err
}

func (p *UserMgmt) DelFriend(param *types.UmDelFriendsParam) (err error) {
	err = p.storage.DelFriend(param)
	return err
}

func (p *UserMgmt) GetFriendList(uid uint64) (friendUidList []uint64, err error) {
	friendUidList, err = p.storage.GetFriendList(uid)
	if err != nil {
		return nil, err
	}
	return friendUidList, nil
}
