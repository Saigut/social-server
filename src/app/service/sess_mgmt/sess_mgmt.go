package sess_mgmt

import (
    "fmt"
    "social_server/src/app/common/types"
    "social_server/src/app/data"
)

type SessMgmt struct {
    storage *data.DB
    cache   *data.Cache
}

func NewSessMgmt(storage *data.DB, cache *data.Cache) *SessMgmt {
    return &SessMgmt{
        storage: storage,
        cache:   cache,
    }
}

func (p *SessMgmt) CreateSess(userName string, timeoutTsS uint64) (sessId types.SessId, err error) {
    // 查询用户id
    user, err := p.storage.GetUserInfoByUsername(userName)
    if err != nil {
        return "", fmt.Errorf("storage.GetUserInfoByUsername: %w", err)
    }
    return p.cache.CreateSess(userName, user.Uid, timeoutTsS)
}

func (p *SessMgmt) GetSessCtx(sessId types.SessId) (sessCtx *types.SessCtx, err error) {
    return p.cache.GetSessCtx(sessId)
}

func (p *SessMgmt) GetSessCtxByUsername(userName string) (sessCtx *types.SessCtx, err error) {
    return p.cache.GetSessCtxByUsername(userName)
}

func (p *SessMgmt) DeleteSess(sessId types.SessId) (err error) {
    return p.cache.DeleteSess(sessId)
}
