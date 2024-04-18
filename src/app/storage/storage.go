package storage

import "errors"

/*

* session:
	- user_login
	- user_logout

* user-mgmt
	- register
	- unregister
	- login
	- logout
	- add_friends
	- del_friends
	- list_friends

* chat
	- get_chat_msg
	- get_chat_msg_hist_with
	- send_chat_msg_to

* video
	- put_video_file
	- get_video_hls

	- get_video_list
	- get_video_metadata
	- get_explorer_video_list
	- get_likes
	- do_like
	- undo_like
	- get_comments
	- add_comment
	- del_comment
*/

type SessUserLoginParam struct {
}

type SessUserLogoutParam struct {
}

type UmRegisterParam struct {
}

type UmUnregisterParam struct {
}

type UmLoginParam struct {
}

type UmLogoutParam struct {
}

type UmAddFriendsParam struct {
}

type UmDelFriendsParam struct {
}

type UmListFriendsParam struct {
}

type ChatGetChatMsgParam struct {
}

type ChatGetChatMsgHistWithParam struct {
}

type ChatSendChatMsgToParam struct {
}

type VideoPutVideoFileParam struct {
}

type VideoGetVideoHlsParam struct {
}

type VideoGetVideoListParam struct {
}

type VideoGetVideoMetadataParam struct {
}

type VideoGetExplorerVideoListParam struct {
}

type VideoGetLikesParam struct {
}

type VideoDoLikeParam struct {
}

type VideoUndoLikeParam struct {
}

type VideoGetCommentsParam struct {
}

type VideoAddCommentParam struct {
}

type VideoDelCommentParam struct {
}

type Storage struct {
}

func (p *Storage) SessUserLogin(param*SessUserLoginParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) SessUserLogout(param *SessUserLogoutParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) UmRegister(param *UmRegisterParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) UmUnregister(param *UmUnregisterParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) UmLogin(param *UmLoginParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) UmLogout(param *UmLogoutParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) UmAddFriends(param *UmAddFriendsParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) UmDelFriends(param *UmDelFriendsParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) UmListFriends(param *UmListFriendsParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) ChatGetChatMsg(param *ChatGetChatMsgParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) ChatGetChatMsgHistWith(param *ChatGetChatMsgHistWithParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) ChatSendChatMsgTo(param *ChatSendChatMsgToParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) VideoPutVideoFile(param *VideoPutVideoFileParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) VideoGetVideoHls(param *VideoGetVideoHlsParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) VideoGetVideoList(param *VideoGetVideoListParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) VideoGetVideoMetadata(param *VideoGetVideoMetadataParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) VideoGetExplorerVideoList(param *VideoGetExplorerVideoListParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) VideoGetLikes(param *VideoGetLikesParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) VideoDoLike(param *VideoDoLikeParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) VideoUndoLike(param *VideoUndoLikeParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) VideoGetComments(param *VideoGetCommentsParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) VideoAddComment(param *VideoAddCommentParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Storage) VideoDelComment(param *VideoDelCommentParam) (err error) {
	return errors.New("method not implemented")
}
