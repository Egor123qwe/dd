package msg

import (
	"encoding/json"
	"fmt"

	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/model"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/model/msg"
)

type Parser interface {
	ParseConnection() (msg.Connection, error)
	WithConnection(dest msg.Connection) (msg.MSG, error)
}

type parser struct {
	data []byte
}

func New(msg msg.MSG) Parser {
	return parser{data: msg.Data}
}

func (p parser) ParseConnection() (msg.Connection, error) {
	var fullMSG msg.Full

	if err := json.Unmarshal(p.data, &fullMSG); err != nil {
		return msg.Connection{}, fmt.Errorf("failed to parse message: %w", err)
	}

	var meta msg.Meta

	if err := json.Unmarshal(fullMSG.Meta, &meta); err != nil {
		return msg.Connection{}, fmt.Errorf("failed to parse message (meta): %w", err)
	}

	switch meta.Conn.Type {
	case msg.UserID:
		if meta.Conn.UserID == "" {
			return msg.Connection{}, model.ErrDestinationNotFound
		}

	case msg.ConnectionID:
		if meta.Conn.ConnID == "" {
			return msg.Connection{}, model.ErrDestinationNotFound
		}

	case msg.AllID:
		if meta.Conn.UserID == "" && meta.Conn.ConnID == "" {
			return msg.Connection{}, model.ErrDestinationNotFound
		}
	}

	return meta.Conn, nil
}

func (p parser) WithConnection(dest msg.Connection) (msg.MSG, error) {
	var fullMSG msg.Full

	if err := json.Unmarshal(p.data, &fullMSG); err != nil {
		return msg.MSG{}, fmt.Errorf("failed to parse message: %w", err)
	}

	var meta map[string]interface{}

	if err := json.Unmarshal(fullMSG.Meta, &meta); err != nil {
		return msg.MSG{}, err
	}

	meta[msg.ConnFieldJSON] = dest

	updatedMeta, err := json.Marshal(meta)
	if err != nil {
		return msg.MSG{}, fmt.Errorf("failed to marshal message (meta): %w", err)
	}

	fullMSG.Meta = updatedMeta

	result, err := json.Marshal(fullMSG)
	if err != nil {
		return msg.MSG{}, fmt.Errorf("failed to marshal message: %w", err)
	}

	return msg.MSG{Data: result}, nil
}
