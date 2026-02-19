package channel

import (
	"context"
	"errors"
	"sync"
)

var ErrChannelClosed = errors.New("channel already closed")

type Channel interface {
	Write(ctx context.Context, msg []byte) error
	Read(ctx context.Context) ([]byte, bool)

	Close() error
}

type channel struct {
	data chan []byte

	lifeCtx context.Context
	close   context.CancelFunc

	mu *sync.Mutex
}

func New(bufferSize int) Channel {
	bufferSize = max(bufferSize, 1)

	ch := channel{
		data: make(chan []byte, bufferSize),

		mu: &sync.Mutex{},
	}

	ch.lifeCtx, ch.close = context.WithCancel(context.Background())

	return ch
}

func (c channel) Write(ctx context.Context, msg []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lifeCtx.Err() != nil {
		return ErrChannelClosed
	}

	select {
	case c.data <- msg:
		return nil

	case <-c.lifeCtx.Done():
		return ErrChannelClosed

	case <-ctx.Done():
		return nil
	}
}

func (c channel) Read(ctx context.Context) ([]byte, bool) {
	select {
	case msg, ok := <-c.data:
		return msg, ok

	case <-ctx.Done():
		return nil, false
	}
}

func (c channel) Close() error {
	if c.lifeCtx.Err() != nil {
		return ErrChannelClosed
	}

	c.close()

	c.mu.Lock()
	defer c.mu.Unlock()

	close(c.data)

	return nil
}
