package service

import (
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/app/server/ws"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/runtimedaemon"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/usecase"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/usecase/state"
)

type Service interface {
	Usecase() usecase.Usecase
	RuntimeDaemon() runtimedaemon.API
}

type service struct {
	usecase usecase.Usecase
	rd      runtimedaemon.API
}

func New(wsConn ws.Conn, rd runtimedaemon.API, token string) (Service, error) {
	srv := service{
		usecase: usecase.New(wsConn, rd, state.New(), token),
		rd:      rd,
	}

	return srv, nil
}

func (s service) RuntimeDaemon() runtimedaemon.API {
	return s.rd
}

func (s service) Usecase() usecase.Usecase {
	return s.usecase
}
