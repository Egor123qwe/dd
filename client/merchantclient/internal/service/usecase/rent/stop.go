package rent

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/event"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg/content/shareP2P"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/usecase/state"
	proto "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/proto/runtimedaemon/generate"
)

func (u usecase) StopEvent(ctx context.Context, sessionID string) error {
	u.state.Mutex().Lock()
	defer u.state.Mutex().Unlock()

	if u.state.GetSessionID() != sessionID {
		return fmt.Errorf("session id is not equal to current session id")
	}

	if err := u.reset(ctx); err != nil {
		log.Warningf("failed to reset merchant: %s", err)
	}

	return nil
}

func (u usecase) StopRequest(ctx context.Context) error {
	u.state.Mutex().Lock()
	defer u.state.Mutex().Unlock()

	if u.state.GetStatus() == state.Disabled {
		return nil
	}

	data := shareP2P.StopMerchantReq{
		SessionID: u.state.GetSessionID(),
	}

	req := msg.Marshal(string(event.ShareP2PStop), msg.Meta{MessageID: uuid.New().String()}, data)

	if err := u.wsConn.Writer().Write(ctx, req); err != nil {
		return fmt.Errorf("failed to send merchant stop request: %w", err)
	}

	if err := u.reset(ctx); err != nil {
		log.Warningf("failed to reset merchant: %s", err)
	}

	return nil
}

func (u usecase) reset(ctx context.Context) error {
	u.state.Reset()
	u.keepAlive.Cancel()
	if u.nodeCheck != nil {
		u.nodeCheck.CancelMerchantNodeCheckLoop()
	}

	u.rd.ChangeMode(ctx, &proto.ChangeModeReq{Mode: proto.Mode_Disable})

	return nil
}
