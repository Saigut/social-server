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

type UmRegisterParam struct {
	Uid      string `json:"uid"`
	Username string `json:"username"`
	Passwd   string `json:"passwd"`
	Email    string `json:"email"`
}

type UmUnregisterParam struct {
	Uid string `json:"uid"`
}

type UmLoginParam struct {
	Uid  string `json:"uid"`
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

type UsermgmtT struct {
}

func (p *UsermgmtT) Register(param*UmRegisterParam) (err error) {
	return errors.New("method not implemented")
}

func (p *UsermgmtT) Unregister(param *UmUnregisterParam) (err error) {
	return errors.New("method not implemented")
}

func (p *UsermgmtT) Login(param *UmLoginParam) (err error) {
	return errors.New("method not implemented")
}

func (p *UsermgmtT) Logout(param *UmLogoutParam) (err error) {
	return errors.New("method not implemented")
}

func (p *UsermgmtT) AddFriends(param *UmAddFriendsParam) (err error) {
	return errors.New("method not implemented")
}

func (p *UsermgmtT) DelFriends(param *UmDelFriendsParam) (err error) {
	return errors.New("method not implemented")
}

func (p *UsermgmtT) ListFriends(param *UmListFriendsParam) (friendsUid []string, err error) {
	return nil, errors.New("method not implemented")
}
