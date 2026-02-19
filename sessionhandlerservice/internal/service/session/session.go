package session

import (
	"context"

	"github.com/op/go-logging"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/session"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/service/rent"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage"
)

var log = logging.MustGetLogger("session")

type IDType string

const (
	SessionIDType IDType = "session_id "
	ClientIDType  IDType = "client_user_id "
	RequestIDType IDType = "request_id "
)

type Service interface {
	Init(ctx context.Context, req session.InitReq) (session.InitResp, error)

	Start(ctx context.Context, req session.StartReq) (session.StartResp, error)
	Stop(ctx context.Context, req session.StopReq) (session.StopResp, error)

	GetClientRents(ctx context.Context, clientID string) ([]string, error)
	GetMerchantRent(ctx context.Context, sessionID string) (string, error)
}

type service struct {
	storage storage.Storage
	rent    rent.Service
	config  config
}

func New(storage storage.Storage, rent rent.Service) Service {
	return &service{
		storage: storage,
		rent:    rent,
		config:  newConfig(),
	}
}

func (s service) GetClientRents(ctx context.Context, clientID string) ([]string, error) {
	return s.storage.Rent().GetClientRents(ctx, clientID)
}

func (s service) GetMerchantRent(ctx context.Context, sessionID string) (string, error) {
	return s.storage.Rent().GetMerchantRent(ctx, sessionID)
}
