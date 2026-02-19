package msg

import (
	"context"
	"fmt"
)

var UnknownEventErr = fmt.Errorf("unknown event")

type TypeParser func(msg []byte) (string, error)
type HandleFunc func(ctx context.Context, msg []byte) error

type Resolver interface {
	Serve(ctx context.Context, msg []byte) error
}

type Router interface {
	Resolver
	Add(event string, fn HandleFunc)
}

type handler struct {
	parser   TypeParser
	handlers map[string]HandleFunc
}

func NewRouter(parser TypeParser) Router {
	return &handler{
		parser:   parser,
		handlers: make(map[string]HandleFunc),
	}
}

func (h *handler) Add(event string, fn HandleFunc) {
	h.handlers[event] = fn
}

func (h *handler) Serve(ctx context.Context, msg []byte) error {
	t, err := h.parser(msg)
	if err != nil {
		return err
	}

	fn, ok := h.handlers[t]
	if !ok {
		return fmt.Errorf("%w: %s", UnknownEventErr, t)
	}

	return fn(ctx, msg)
}
