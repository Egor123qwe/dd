package app

import (
	"fmt"
	"net/http"

	"github.com/spf13/viper"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/config"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/ws"
)

func newConn(token string) (ws.Conn, error) {
	return newConnWithURL(token, "")
}

func newConnWithURL(token string, connectionURL string) (ws.Conn, error) {
	header := http.Header{}
	header.Set("Client-Authorization", fmt.Sprintf("Bearer %s", token))

	url := connectionURL
	if url == "" {
		url = viper.GetString(config.SmConnectionUrlKey)
	}

	conn, err := ws.Connect(url, header)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return conn, nil
}
