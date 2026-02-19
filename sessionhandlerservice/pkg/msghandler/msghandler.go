package msghandler

import (
	"context"
)

type EventParser func(msg []byte) (string, error)
type HandleFunc func(ctx context.Context, msg []byte) error

type Server interface {
	Serve(ctx context.Context, msg []byte) error
}

type MsgHandler interface {
	Server
	Add(event string, fn HandleFunc)
}

type handler struct {
	eventParser EventParser
	handlers    map[string]HandleFunc
}

func New(parser EventParser) MsgHandler {
	return &handler{
		eventParser: parser,
		handlers:    make(map[string]HandleFunc),
	}
}

func (h *handler) Add(event string, fn HandleFunc) {
	h.handlers[event] = fn
}

func (h *handler) Serve(ctx context.Context, msg []byte) error {
	event, err := h.eventParser(msg)
	if err != nil {
		return err
	}

	fn, ok := h.handlers[event]
	if !ok {
		return nil
	}

	return fn(ctx, msg)
}
