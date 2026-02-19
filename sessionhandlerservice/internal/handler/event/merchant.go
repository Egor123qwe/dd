package event

import (
	"context"
	"errors"
	"fmt"

	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/handler/model/msg"
	msgcontent "gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/handler/model/msg/content"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model"
	rentRepo "gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/db/psql/repo/rent"
)

func (h handler) MerchantStopped(ctx context.Context, m []byte) error {
	reqMSG, err := msg.New(m).Parse()
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	shareP2PStopContent := new(msgcontent.ShareP2PStopContent)

	if err := reqMSG.ParseContent(&shareP2PStopContent); err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	requestID, err := h.srv.Session().GetMerchantRent(ctx, shareP2PStopContent.SessionID)
	if err != nil {
		if errors.Is(err, rentRepo.ErrRentNotFound) {
			// Сессия создана (resourcepool), но аренда не была начата (нет записи rent) — ничего останавливать
			log.Infof("merchant stopped for session %s: rent not created yet, skipping stop", shareP2PStopContent.SessionID)
			return nil
		}
		return fmt.Errorf("%w: %w", model.ErrFailedToGetMerchantRent, err)
	}

	content := msgcontent.StopSessionReq{RequestID: requestID}

	handlerReq := createResp(reqMSG.Type, reqMSG.Meta, content)

	if err := h.StopSession(ctx, handlerReq); err != nil {
		log.Errorf("failed to complite rent stop event from merchant stopped: %s", err)
	}

	return nil
}
