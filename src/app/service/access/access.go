package access

// http server
// authentication using http, if passed, upgrade to websocket
// forward protocol buffer between `access` and `service` api.


type SessInfo struct {
	SessId uint64
	SessToken string
	Uid string
}

type SessUserLoginParam struct {
	Uid string
	Passwd string
}

type SessUserLogoutParam struct {
	Uid string
	SessToken string
}

type GetUserSessListParam struct {
	Uid string
}

type FindUserSessParam struct {
	Uid string
	SessToken string
}

