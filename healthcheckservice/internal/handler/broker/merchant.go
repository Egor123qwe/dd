package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/handler/model/message"
)

func (h handler) HandleMerchantStartRent(ctx context.Context, msg []byte) error {
	const op = "HandleMerchantStartRent"
	var MerchantReq message.MerchantRent

	err := json.Unmarshal(msg, &MerchantReq)

	if err != nil {
		return fmt.Errorf("%s:%v", op, err)
	}

	err = h.service.Status().RentMerchant(ctx, MerchantReq)
	if err != nil {
		return err
	}

	return nil
}
