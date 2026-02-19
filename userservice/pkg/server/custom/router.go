package custom

import (
	"context"
	"fmt"
)

var UnknownRouteErr = fmt.Errorf("unknown route")

type (
	RouteParserFn func(msg []byte) (string, error)
	HandleFn      func(ctx context.Context, msg []byte) error
)

type Router interface {
	Resolver
	Add(event string, fn HandleFn)
}

type Resolver interface {
	Resolve(ctx context.Context, msg []byte) error
}

type Handler interface {
	Handle(ctx context.Context, msg []byte) error
}

type router struct {
	parser   RouteParserFn
	handlers map[string]HandleFn
}

func NewRouter(parser RouteParserFn) Router {
	return &router{
		parser:   parser,
		handlers: make(map[string]HandleFn),
	}
}

func (h *router) Add(event string, fn HandleFn) {
	h.handlers[event] = fn
}

func (h *router) Resolve(ctx context.Context, msg []byte) error {
	t, err := h.parser(msg)
	if err != nil {
		return err
	}

	fn, ok := h.handlers[t]
	if !ok {
		return fmt.Errorf("%w: %s", UnknownRouteErr, t)
	}

	return fn(ctx, msg)
}
