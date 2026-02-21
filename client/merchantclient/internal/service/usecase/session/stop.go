package session

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/event"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg/content/session"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/usecase/state"
	proto "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/proto/runtimedaemon/generate"
)

func (u usecase) StopEvent(ctx context.Context, requestID string) error {
	u.startController.Cancel()

	u.state.Mutex().Lock()
	defer u.state.Mutex().Unlock()

	if u.state.GetStatus() != state.InRent {
		return nil
	}

	if u.state.GetRequestID() != requestID {
		return fmt.Errorf("request id is not equal to current request id")
	}

	// Покупатель остановил аренду — переводим поставщика обратно в «Готов к аренде» (Ready),
	// не отключая сессию и не вызывая ChangeMode(Disable), чтобы поставщик оставался в списке и мог принять следующего покупателя.
	u.transitionToReadyAfterRentEnd()

	log.Infof("session [%s] stopped by client, merchant ready for next rent", requestID)

	return nil
}

// TransitionToReadyAfterRentEnd вызывается при получении share-p2p-stop: аренда завершена, переводим в Ready без отключения сессии.
func (u usecase) TransitionToReadyAfterRentEnd(ctx context.Context, sessionID string) error {
	u.state.Mutex().Lock()
	defer u.state.Mutex().Unlock()

	if u.state.GetStatus() != state.InRent {
		return nil
	}
	if u.state.GetSessionID() != sessionID {
		return fmt.Errorf("session id mismatch")
	}

	u.transitionToReadyAfterRentEnd()
	log.Infof("session [%s] rent ended (share-p2p-stop), merchant ready for next rent", sessionID)
	return nil
}

func (u usecase) StopRequest(ctx context.Context, reason string) error {
	u.startController.Cancel()

	u.state.Mutex().Lock()
	defer u.state.Mutex().Unlock()

	if u.state.GetStatus() != state.InRent {
		return nil
	}

	data := session.SessionStopReq{
		RequestID: u.state.GetRequestID(),
		Reason:    reason,
	}

	req := msg.Marshal(string(event.StopSession), msg.Meta{MessageID: uuid.New().String()}, data)

	if err := u.wsConn.Writer().Write(ctx, req); err != nil {
		return fmt.Errorf("failed to send merchant stop request: %w", err)
	}

	requestID := u.state.GetRequestID()

	if err := u.reset(ctx); err != nil {
		log.Warningf("failed to reset merchant: %s", err)
	}

	log.Infof("session [%s] stoped by reason: %s", requestID, reason)

	return nil
}

// stopNodeFromPortal вызывается, когда узел отключили с портала: для InRent — отправляем stop координатору и сбрасываем, для Ready — только сброс (как «завершить сеанс»).
func (u usecase) stopNodeFromPortal(ctx context.Context) error {
	u.state.Mutex().Lock()
	st := u.state.GetStatus()
	u.state.Mutex().Unlock()

	if st != state.InRent && st != state.Ready {
		return nil
	}
	if st == state.InRent {
		return u.StopRequest(ctx, "merchant node stopped via portal (HTTP check)")
	}
	// Ready: узел был «готов к приёму», сбрасываем без отправки stop (аренды нет)
	u.state.Mutex().Lock()
	defer u.state.Mutex().Unlock()
	return u.reset(ctx)
}

// transitionToReadyAfterRentEnd переводит мерчанта из InRent в Ready после окончания аренды (например, покупатель остановил).
// Сессия и узел остаются активными, daemon не отключается.
func (u usecase) transitionToReadyAfterRentEnd() {
	u.state.SetStatus(state.Ready)
	u.state.SetRequestID("")
	u.state.SetRentStartedAt(nil)

	u.sessionHealthCheck.Cancel()
	u.sessionStatusHTTPCheck.Cancel()
}

func (u usecase) reset(ctx context.Context) error {
	u.state.SetStatus(state.Disabled)
	u.state.SetSessionID("")
	u.state.SetRequestID("")
	u.state.SetRentStartedAt(nil)
	u.currentSessionID = ""

	u.sessionHealthCheck.Cancel()
	u.sessionStatusHTTPCheck.Cancel()
	u.merchantNodeHTTPCheck.Cancel()

	if _, err := u.rd.ChangeMode(ctx, &proto.ChangeModeReq{Mode: proto.Mode_Disable}); err != nil {
		return fmt.Errorf("failed to change mode: %w", err)
	}

	return nil
}
