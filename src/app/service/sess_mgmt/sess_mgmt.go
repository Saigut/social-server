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

func (p *SessMgmt) CreateSess(username string, uid uint64, expireAfterSecs uint64) (sessId types.SessId, err error) {
    sessId, err = p.cache.CreateSess(username, uid, expireAfterSecs)
    if err != nil {
        return "", fmt.Errorf("cache.CreateSess: %w", err)
    }
    return sessId, nil
}

func (p *SessMgmt) GetSessCtx(sessId types.SessId) (sessCtx *types.SessCtx, err error) {
    return p.cache.GetSessCtx(sessId)
}

func (p *SessMgmt) GetSessCtxByUid(uid uint64) (sessCtx *types.SessCtx, err error) {
    return p.cache.GetSessCtxByUid(uid)
}

func (p *SessMgmt) RenewSessCtx(sessCtx *types.SessCtx, expireAfterSecs uint64) error {
    return p.cache.RenewSessCtx(sessCtx, expireAfterSecs)
}

func (p *SessMgmt) DeleteSess(sessId types.SessId) (err error) {
    return p.cache.DeleteSess(sessId)
}

func (p *SessMgmt) DeleteUserSess(uid uint64) (err error) {
    return p.cache.DeleteUserSess(uid)
}
