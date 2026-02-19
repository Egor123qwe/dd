package event

import (
	"github.com/op/go-logging"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/handler/event/session"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/handler/event/shareP2P"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/event"
	msgModel "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/server/msg"
)

var log = logging.MustGetLogger("handler")

type handler struct {
	srv    service.Service
	router msg.Router
}

func New(srv service.Service) msg.Resolver {
	handler := &handler{
		srv:    srv,
		router: msg.NewRouter(msgModel.NewTypeParser()),
	}

	handler.initEvents()

	return handler.router
}

func (h handler) initEvents() {
	shareP2P := shareP2P.New(h.srv.Usecase())
	session := session.New(h.srv.Usecase())

	h.router.Add(string(event.ShareP2PStop), shareP2P.Stop)
	h.router.Add(string(event.ExpiredSession), shareP2P.Stop)

	h.router.Add(string(event.MerchantStartRentEvent), session.Start)
	h.router.Add(string(event.SessionStatusUpdatedEvent), session.StatusUpdated)
}
