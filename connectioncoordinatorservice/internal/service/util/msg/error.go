package msg

import (
	"encoding/json"

	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/model/msg"
)

func ErrorResponse(err error) msg.MSG {
	resp := msg.ErrorResp{Error: err.Error()}

	data, _ := json.Marshal(resp)

	return msg.MSG{Data: data}
}
