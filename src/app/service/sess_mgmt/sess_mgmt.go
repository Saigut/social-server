package sess_mgmt

import "errors"

/*
	1. 创建会话
	2. 查询会话
	3. 删除会话
*/

type SessIdT string

type SessCtxT struct {
	SessId SessIdT
	Uid    string
}

type SessMgmtT struct {
}

func (p *SessMgmtT) CreateSess(uid string, timeoutTsS uint64) (sessId *SessIdT, err error) {
	return nil, errors.New("method not implemented")
}

func (p *SessMgmtT) GetSessCtx(sessId SessIdT) (sessCtx *SessCtxT, err error) {
	return nil, errors.New("method not implemented")
}

func (p *SessMgmtT) GetSessCtxByUsername(userName string) (sessCtx *SessCtxT, err error) {
	return nil, errors.New("method not implemented")
}

func (p *SessMgmtT) DeleteSess(sessId SessIdT) (err error) {
	return errors.New("method not implemented")
}
