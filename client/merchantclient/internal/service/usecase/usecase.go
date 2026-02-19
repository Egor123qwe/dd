package usecase

import (
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/app/server/ws"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/runtimedaemon"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/usecase/rent"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/usecase/session"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/usecase/state"
)

type Usecase interface {
	Rent() rent.Usecase
	Session() session.Usecase
}

type usecase struct {
	shareP2P rent.Usecase
	session  session.Usecase
}

func New(wsConn ws.Conn, rd runtimedaemon.API, state state.State, token string) Usecase {
	sess := session.New(wsConn, rd, state, token)
	return &usecase{
		shareP2P: rent.New(wsConn, rd, state, sess),
		session:  sess,
	}
}

func (u *usecase) Rent() rent.Usecase {
	return u.shareP2P
}

func (u *usecase) Session() session.Usecase {
	return u.session
}
