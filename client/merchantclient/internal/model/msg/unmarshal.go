package msg

import (
	"encoding/json"
	"fmt"

	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/server/msg"
)

func Unmarshal(data []byte) (MSG, error) {
	var msg MSG

	if err := json.Unmarshal(data, &msg); err != nil {
		return MSG{}, fmt.Errorf("failed to parse message: %w", err)
	}

	return msg, nil
}

func (m MSG) UnmarshalContent(dest any) error {
	return json.Unmarshal(m.Content, dest)
}

func NewTypeParser() msg.TypeParser {
	parser := func(m []byte) (string, error) {
		msg, err := Unmarshal(m)

		return msg.Type, err
	}

	return parser
}

func NewIDParser() msg.IdParser {
	parser := func(m []byte) (string, error) {
		msg, err := Unmarshal(m)

		return msg.Meta.MessageID, err
	}

	return parser
}
