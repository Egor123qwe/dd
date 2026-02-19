package listener

import (
	"context"
	"errors"
	"fmt"

	"github.com/op/go-logging"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/broker"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/model/msg"
	parser "gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/service/util/msg"
)

var log = logging.MustGetLogger("listener")

var (
	ErrConnListenerClosed = errors.New("connection listener closed")
)

type Listener interface {
	Subscribe(id string, userID string) ConnListener
}

type listener struct {
	broker      broker.Service
	connections connections
}

func NewListener(broker broker.Service) Listener {
	listener := &listener{
		broker:      broker,
		connections: newState(),
	}

	go listener.listen(context.Background())

	return listener
}

func (l listener) Subscribe(connectionID string, userID string) ConnListener {
	unsubscribe := func() {
		l.connections.remove(connectionID)
	}

	conn := newConnListener(unsubscribe)

	l.connections.add(userID, connectionID, conn)

	return conn
}

func (l listener) listen(ctx context.Context) {
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		message, err := l.broker.Consumer().Consume(ctx)
		if err != nil {
			log.Errorf("failed to consume message: %v", err)

			continue
		}

		go func() {
			if err := l.send(message); err != nil {
				log.Errorf("failed to send message: %v", err)
			}
		}()
	}
}

func (l listener) send(m msg.MSG) error {
	dest, err := parser.New(m).ParseConnection()
	if err != nil {
		log.Errorf("output message: failed to parse connection: %v", err)
		return fmt.Errorf("failed to parse connection id: %v", err)
	}

	var listeners []connListener

	switch dest.Type {
	case msg.ConnectionID:
		conn, ok := l.connections.get(dest.ConnID)
		if !ok {
			log.Errorf("output message: connection not found for conn_id=%q (message not delivered to client)", dest.ConnID)
			return fmt.Errorf("failed to load connection: connection not found")
		}

		listeners = append(listeners, conn)

	case msg.UserID:
		listeners = l.connections.getAll(dest.UserID)
		if len(listeners) == 0 {
			return fmt.Errorf("failed to load connections: connections not found")
		}

	case msg.AllID:
		return fmt.Errorf("invalid type of destination")
	}

	// broadcast
	for _, listener := range listeners {
		go func() { listener.msgCh <- m }()
	}

	return nil
}
