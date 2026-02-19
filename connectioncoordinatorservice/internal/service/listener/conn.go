package listener

import (
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/model/msg"
)

type ConnListener interface {
	Read() (msg.MSG, error)
	Close()
}

type connListener struct {
	msgCh   chan msg.MSG
	closeFn func()
}

func newConnListener(unsubscribeFn func()) connListener {
	msgCh := make(chan msg.MSG)

	conn := connListener{
		msgCh: msgCh,

		closeFn: func() {
			unsubscribeFn()
			close(msgCh)
		},
	}

	return conn
}

func (c connListener) Close() {
	c.closeFn()
}

func (c connListener) Read() (msg.MSG, error) {
	m, ok := <-c.msgCh
	if !ok {
		return msg.MSG{}, ErrConnListenerClosed
	}

	return m, nil
}
