package rent

import (
	"context"
	"fmt"

	"github.com/op/go-logging"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/rent"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage"
)

var log = logging.MustGetLogger("rent")

type Service interface {
	GenerateRentSettings(ctx context.Context, req rent.RequestSettings) (rent.Settings, error)
	GetRentSettings(ctx context.Context, requestID string) (rent.Settings, error)
}

type service struct {
	storage storage.Storage
	config  config
}

func New(storage storage.Storage) Service {
	return &service{
		config:  newConfig(),
		storage: storage,
	}
}

func (s service) GenerateRentSettings(ctx context.Context, req rent.RequestSettings) (rent.Settings, error) {
	template, err := s.template(ctx, req.TemplateID)
	if err != nil {
		return rent.Settings{}, fmt.Errorf("failed to get template: %w", err)
	}

	network, err := s.network(ctx, req.Mode, template.Template.Ports)
	if err != nil {
		return rent.Settings{}, fmt.Errorf("failed to get network: %w", err)
	}

	result := rent.Settings{
		Mode:     req.Mode,
		Template: template,
		Network:  network,
	}

	return result, nil
}

func (s service) GetRentSettings(ctx context.Context, requestID string) (rent.Settings, error) {
	return s.storage.Rent().GetSettings(ctx, requestID)
}
