package ws

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var ErrDisconnected = errors.New("disconnected from server")

type Conn interface {
	// Reader returns a channel of read messages
	Reader() Reader

	// Writer returns a channel to Write messages
	Writer() Writer

	// Close closes the connection
	Close() error
}

type conn struct {
	conn *websocket.Conn
}

func Connect(url string, header http.Header) (Conn, error) {
	c, resp, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		if resp != nil {
			errInfoData, _ := io.ReadAll(resp.Body)
			errInfo := string(errInfoData)

			if len(strings.Split(errInfo, ":")) > 0 {
				errInfo = strings.Split(errInfo, ":")[0]
			}

			return nil, fmt.Errorf("%w. Error info: %s", err, errInfo)
		}

		return nil, err
	}

	return New(c), nil
}

func New(c *websocket.Conn) Conn {
	return &conn{conn: c}
}

func (c conn) Close() error {
	return c.conn.Close()
}
