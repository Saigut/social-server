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
	Uid string `json:"uid"`
}

type UmLoginParam struct {
	Uid    string `json:"uid"`
	Passwd string `json:"passwd"`
}

type UmLogoutParam struct {
	Uid string `json:"uid"`
}

type UmAddFriendsParam struct {
	Uid        string   `json:"uid"`
	FriendsUid []string `json:"friends_uid"`
}

type UmDelFriendsParam struct {
	Uid        string   `json:"uid"`
	FriendsUid []string `json:"friends_uid"`
}

type UmListFriendsParam struct {
	Uid        string   `json:"uid"`
	FriendsUid []string `json:"friends_uid"`
}

type UserMgmtT struct {
}

func (p *UserMgmtT) UserInfoValidate(param *UmUserInfoValidateParam) (err error) {
	return errors.New("method not implemented")
}

func (p *UserMgmtT) IsUsernameExisted(param *UmIsUsernameExistedParam) (err error) {
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

func (p *UserMgmtT) ListFriends(param *UmListFriendsParam) (friendsUid []string, err error) {
	return nil, errors.New("method not implemented")
}
