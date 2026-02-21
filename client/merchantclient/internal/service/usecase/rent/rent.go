package rent

import (
	"context"
	"time"

	"github.com/op/go-logging"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/app/server/ws"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/event"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/runtimedaemon"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/usecase/state"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/sync/fnController"
)

const (
	// Отправлять чаще, чем TTL сессии (1 мин на бэкенде), чтобы ключ не истекал при задержках
	keepAliveInterval = 25 * time.Second
)

var log = logging.MustGetLogger("shareP2P")

// SessionNodeChecker запуск/остановка проверки «узел в списке активных».
type SessionNodeChecker interface {
	StartMerchantNodeCheckLoop(ctx context.Context)
	CancelMerchantNodeCheckLoop()
}

type Usecase interface {
	Init(ctx context.Context, cheatMode bool, nodeName string, memoryLimitBytes, storageLimitBytes, cpuLimit int64) error

	StopRequest(ctx context.Context) error
	StopEvent(ctx context.Context, sessionID string) error

	GetStatus() state.Status
	GetRentStartedAt() *time.Time
	GetTotalPrice() float64
}

type usecase struct {
	wsConn    ws.Conn
	state     state.State
	rd        runtimedaemon.API
	nodeCheck SessionNodeChecker
	keepAlive fnController.Controller
}

func New(wsConn ws.Conn, rd runtimedaemon.API, state state.State, nodeCheck SessionNodeChecker) Usecase {
	return &usecase{
		wsConn:    wsConn,
		state:     state,
		rd:        rd,
		nodeCheck: nodeCheck,
		keepAlive: fnController.New(),
	}
}

func (u usecase) GetStatus() state.Status {
	return u.state.GetStatus()
}

func (u usecase) GetRentStartedAt() *time.Time {
	return u.state.GetRentStartedAt()
}

func (u usecase) GetTotalPrice() float64 {
	return u.state.GetTotalPrice()
}

func (u usecase) KeepAlive(ctx context.Context) {
	log.Infof("merchant keep-alive started")

	req := msg.Marshal(string(event.KeepAlive), msg.Meta{}, struct{}{})

	for {
		select {
		case <-ctx.Done():
			log.Infof("merchant keep-alive stopped")
			return

		case <-time.After(keepAliveInterval):
			if err := u.wsConn.Writer().Write(ctx, req); err != nil {
				log.Warningf("failed to send keep-alive: %s", err)
			}

			log.Debugf("keep-alive messgage was send")
		}
	}

}
