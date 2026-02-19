package shareP2P

import (
	"context"

	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/usecase"
)

type Handler interface {
	Stop(ctx context.Context, m []byte) error
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
