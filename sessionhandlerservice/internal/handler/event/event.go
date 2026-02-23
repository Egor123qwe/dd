package event

import (
	"context"

	"github.com/op/go-logging"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/broker/kafka"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/broker/kafka/producer"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/handler/model/msg"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/handler/model/msg/event"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/service"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/pkg/msghandler"
)

var log = logging.MustGetLogger("content handler")

type handler struct {
	srv    service.Service
	router msghandler.MsgHandler

	respondent respondent
}

type respondent struct {
	ws producer.Producer
}

func New(srv service.Service, broker kafka.Service) msghandler.Server {
	config := newConfig()

	// type parser for messages of our contract type
	eventParser := func(m []byte) (string, error) {
		msg, err := msg.New(m).Parse()

		return msg.Type, err
	}

	handler := &handler{
		srv:    srv,
		router: msghandler.New(eventParser),

		respondent: respondent{
			ws: broker.Producer(config.ws.topic),
		},
	}

	handler.initEvents()

	return handler.router
}

func (h handler) initEvents() {
	h.router.Add(string(event.StartSessionEvent), h.InitSession)
	h.router.Add(string(event.StopSessionEvent), h.StopSession)

	// Временно отключено: завершение сессии по expired (rent/client/session). Чтобы включить — заменить IgnoreExpired на h.StopSession / h.MerchantStopped / h.ClientExpired.
	h.router.Add(string(event.ExpiredRentEvent), h.IgnoreExpired)
	h.router.Add(string(event.ExpiredSessionEvent), h.IgnoreExpired)
	h.router.Add(string(event.ExpiredClientEvent), h.IgnoreExpired)

	h.router.Add(string(event.ShareP2PStop), h.MerchantStopped)

	h.router.Add(string(event.RentRequestStatusUpdatedEvent), h.UpdateRentStatus)
}

// IgnoreExpired — временный no-op: не завершаем сессию по причине expired (rent/client/session).
func (h handler) IgnoreExpired(ctx context.Context, _ []byte) error {
	return nil
}
