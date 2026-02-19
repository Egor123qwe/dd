package session

import (
	"context"

	"github.com/op/go-logging"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/usecase"
)

var log = logging.MustGetLogger("session")

type Handler interface {
	Start(ctx context.Context, m []byte) error

	StatusUpdated(ctx context.Context, m []byte) error
}

type handler struct {
	usecase usecase.Usecase
}

func New(usecase usecase.Usecase) Handler {
	handler := &handler{
		usecase: usecase,
	}

	return handler
}
