package user_mgmt

import "errors"

/*
* user-mgmt
	- register
	- unregister
	- login
	- logout
	- add_friends
	- del_friends
	- list_friends
*/

type UmUserInfoValidateParam struct {
	Username  string
	Passphase string
	Email     string
}

type UmIsUsernameExistedParam struct {
	Username string
}

type UmUserAuthenticateParam struct {
	Username  string
	Passphase string
}

type UmRegisterParam struct {
	Username string `json:"username"`
	Passwd   string `json:"passwd"`
	Email    string `json:"email"`
}

type UmUnregisterParam struct {
	Uid uint64 `json:"uid"`
}

type UmLoginParam struct {
	Uid    uint64 `json:"uid"`
	Passwd string `json:"passwd"`
}

type UmLogoutParam struct {
	Uid uint64 `json:"uid"`
}

type UmAddFriendsParam struct {
	Uid        uint64
	FriendUid  uint64
}

type UmDelFriendsParam struct {
	Uid        uint64
	FriendUid  uint64
}

type UmListFriendsParam struct {
	Uid        uint64   `json:"uid"`
	FriendsUid []string `json:"friends_uid"`
}

type UserMgmtT struct {
}

func (p *UserMgmtT) UserInfoValidate(param *UmUserInfoValidateParam) (err error) {
	return errors.New("method not implemented")
}

func (p *UserMgmtT) IsUsernameExisted(userName string) (err error) {
	return errors.New("method not implemented")
}

func (p *UserMgmtT) UserAuthenticate(param *UmUserAuthenticateParam) (err error) {
	return errors.New("method not implemented")
}

func (p *UserMgmtT) Register(param *UmRegisterParam) (err error) {
	return errors.New("method not implemented")
}

func (p *UserMgmtT) Unregister(param *UmUnregisterParam) (err error) {
	return errors.New("method not implemented")
}

func (p *UserMgmtT) Login(param *UmLoginParam) (err error) {
	return errors.New("method not implemented")
}

func (p *UserMgmtT) Logout(param *UmLogoutParam) (err error) {
	return errors.New("method not implemented")
}

func (p *UserMgmtT) AddFriends(param *UmAddFriendsParam) (err error) {
	return errors.New("method not implemented")
}

func (p *UserMgmtT) DelFriends(param *UmDelFriendsParam) (err error) {
	return errors.New("method not implemented")
}

func (p *UserMgmtT) GetFriendList(uid uint64) (friendsUid []uint64, err error) {
	return nil, errors.New("method not implemented")
}
