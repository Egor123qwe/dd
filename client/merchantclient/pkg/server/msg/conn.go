package msg

import (
	"context"
	"errors"
	"fmt"

	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/sync/safeMap"
)

var ErrUnknownResp = fmt.Errorf("unknown response")
var ErrRequestCanceled = errors.New("request canceled")

type IdParser func(msg []byte) (string, error)
type RequestFn func(ctx context.Context, msg []byte) error

type Conn interface {
	Do(ctx context.Context, fn RequestFn, msg []byte) ([]byte, error)
	Serve(msg []byte) error
}

type conn struct {
	parser   IdParser
	requests *safeMap.Map[string, chan []byte]
}

func NewConn(parser IdParser) Conn {
	return &conn{
		parser:   parser,
		requests: safeMap.New[string, chan []byte](),
	}
}

func (c *conn) Do(ctx context.Context, fn RequestFn, msg []byte) ([]byte, error) {
	id, err := c.parser(msg)
	if err != nil {
		return nil, err
	}

	respCh := make(chan []byte, 1)

	c.requests.Add(id, respCh)
	defer close(respCh)
	defer c.requests.Delete(id)

	if err := fn(ctx, msg); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ErrRequestCanceled

	case resp := <-respCh:
		return resp, nil
	}
}

func (c *conn) Serve(msg []byte) error {
	id, err := c.parser(msg)
	if err != nil {
		return err
	}

	ch, ok := c.requests.Get(id)
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownResp, id)
	}

	ch <- msg

	return nil
}
