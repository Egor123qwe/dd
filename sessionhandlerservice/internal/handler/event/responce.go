package event

import (
	"encoding/json"

	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/handler/model/msg"
)

func createErrorResp(t string, meta msg.Meta, err error) []byte {
	meta.Err = &msg.Err{
		Code:    "err",
		Message: err.Error(),
	}
	meta.Status = string(msg.Error)

	return createResp(t, meta, struct{}{})
}

func createResp(t string, meta msg.Meta, content any) []byte {
	m := msg.MSG{
		Type: t,
		Meta: meta,
	}

	m.Content, _ = json.Marshal(content)
	result, _ := json.Marshal(m)

	return result
}
