package post

import "errors"

/*
* post
	stream
	- put_post
	- get_video_hls
	msg
	- get_post_list
	- get_post_metadata
	- get_explorer_post_list
	- get_likes
	- do_like
	- undo_like
	- get_comments
	- add_comment
	- del_comment
*/

type PostMetadata struct {
	PostID uint64
	Uid string
	TsMs uint64
	Des string
	VideoHlsList []string
}

type PostComment struct {
	TsMs uint64
	Content []string
}

type PutPostParam struct {
	Uid string
	TsMs uint64
	FileBuf []byte
	Des string
}

type GetVideoHlsParam struct {
	PostID uint64
	HlsFileName string
}

type GetPostListParam struct {
	Uid string
}

type GetPostMetadataParam struct {
	PostID uint64
}

type GetExplorerVideoListParam struct {
}

type GetLikesParam struct {
	PostID uint64
}

type DoLikeParam struct {
	PostID uint64
	Uid string
}

type UndoLikeParam struct {
	PostID uint64
	Uid string
}

type GetCommentsParam struct {
	PostID uint64
}

type AddCommentParam struct {
	PostID  uint64
	Comment PostComment
}

type DelCommentParam struct {
	PostID uint64
	CommentId uint64
}

type Post struct {
	
}

func (p *Post) PutPost(param *PutPostParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Post) GetVideoHls(param *GetVideoHlsParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Post) GetPostList(param *GetPostListParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Post) GetPostMetadata(param *GetPostMetadataParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Post) GetExplorerPostList(param *GetExplorerVideoListParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Post) GetLikes(param *GetLikesParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Post) DoLike(param *DoLikeParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Post) UndoLike(param *UndoLikeParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Post) GetComments(param *GetCommentsParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Post) AddComment(param *AddCommentParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Post) DelComment(param *DelCommentParam) (err error) {
	return errors.New("method not implemented")
}