package msg

import (
	"encoding/json"
	"fmt"
)

type Parser interface {
	Parse() (MSG, error)
}

type parser struct {
	data []byte
}

func New(msg []byte) Parser {
	return parser{data: msg}
}

func (p parser) Parse() (MSG, error) {
	var msg MSG
	
	if err := json.Unmarshal(p.data, &msg); err != nil {
		return MSG{}, fmt.Errorf("failed to parse message: %w", err)
	}

	return msg, nil
}

func (m MSG) MSG() MSG {
	return m
}

func (m MSG) ParseContent(dest any) error {
	return json.Unmarshal(m.Content, dest)
}
