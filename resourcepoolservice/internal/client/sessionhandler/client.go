package sessionhandler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Client вызывает sessionhandlerservice для остановки аренды по session_id узла.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// StopRentByMerchantSessionID отправляет POST /internal/merchant/session/:sessionId/stop.
func (c *Client) StopRentByMerchantSessionID(ctx context.Context, sessionID string) error {
	url := c.baseURL + "/internal/merchant/session/" + sessionID + "/stop"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("session handler returned %d", resp.StatusCode)
	}
	return nil
}
