package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/config"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/app/server/ws"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg/content/session"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/runtimedaemon"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/usecase/state"
	proto "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/proto/runtimedaemon/generate"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/sync/fnController"
)

var log = logging.MustGetLogger("session")

const (
	healthCheckInterval = 5 * time.Second

	// used to give runtime daemon some time to restart rent in case of error
	healthCheckMaxErrCount = 3

	// Интервал HTTP-проверки: сессия ещё активна на бэкенде или уже выключена
	statusCheckHTTPInterval = 5 * time.Second
)

type Usecase interface {
	Start(ctx context.Context, req session.MerchantRentStartReq) error

	StopRequest(ctx context.Context, reason string) error
	StopEvent(ctx context.Context, requestID string) error

	// StartMerchantNodeCheckLoop запускает фоновую проверку «узел ещё в списке активных» (для Ready и InRent). Вызывается из rent.Init().
	StartMerchantNodeCheckLoop(ctx context.Context)
	// CancelMerchantNodeCheckLoop останавливает проверку (вызывается из rent при «завершить сеанс»).
	CancelMerchantNodeCheckLoop()
}

type usecase struct {
	wsConn ws.Conn

	state state.State
	rd    runtimedaemon.API

	// token для HTTP-проверки статуса сессии (Bearer)
	token string

	// currentSessionID сохраняем при Start(), чтобы HTTP-проверка не зависела от state
	currentSessionID string

	sessionHealthCheck     fnController.Controller
	sessionStatusHTTPCheck fnController.Controller
	merchantNodeHTTPCheck  fnController.Controller
	startController        fnController.Controller
}

func New(wsConn ws.Conn, rd runtimedaemon.API, state state.State, token string) Usecase {
	u := &usecase{
		wsConn: wsConn,

		state: state,
		rd:    rd,
		token: token,

		sessionHealthCheck:     fnController.New(),
		sessionStatusHTTPCheck: fnController.New(),
		merchantNodeHTTPCheck:  fnController.New(),
		startController:        fnController.New(),
	}

	return u
}

func (u usecase) HealthCheck(ctx context.Context) {
	stream, err := u.rd.GetStateStream(ctx, &proto.RuntimeStateStreamReq{})
	if err != nil {
		log.Fatalf("falid to connect to daemon state stream: %v", err)
	}

	log.Infof("merchant session health check started")
	defer log.Infof("merchant session health check stopped")

	defer stream.CloseSend()
	errCounter := 0

	for {
		select {
		case <-ctx.Done():
			return

		case <-time.After(healthCheckInterval):
		}

		data, err := stream.Recv()
		if err != nil {
			status := fmt.Sprintf("daemon worker falid in state stream: %s", err)

			if err := u.StopRequest(context.Background(), status); err != nil {
				log.Errorf("falid to stop session: %s", err)
			}

			// this hack with suffix is needed to avoid grpc context errors
			if !errors.Is(err, context.Canceled) && !strings.HasSuffix(err.Error(), context.Canceled.Error()) {
				log.Errorf(status)
			}

			return
		}

		if data.Mode == proto.Mode_Disable || data.Status != proto.RuntimeState_Ok {
			errCounter++

			status := fmt.Sprintf(
				"daemon worker falid in rent (%d). Status: %s. Mode: %s. %s",
				errCounter, data.Status, data.Mode, data.StatusMsg,
			)

			if errCounter >= healthCheckMaxErrCount {
				if err := u.StopRequest(context.Background(), status); err != nil {
					log.Errorf("falid to stop session: %s", err)
				}

				log.Errorf(status)
				return
			}

			log.Warningf(status)

			continue
		}

		errCounter = 0
	}
}

// StatusCheckLoop периодически запрашивает по HTTP, активна ли ещё текущая сессия (не выключил ли клиент аренду).
// Мерчант-клиент ставится локально у пользователей и стучится на сервер извне — URL должен указывать на ваш бэкенд (healthcheckservice).
// Берётся из config status_check.url или env STATUS_CHECK_URL (например https://api.example.com). Если пусто — проверка отключена.
func (u usecase) StatusCheckLoop(ctx context.Context) {
	baseURL := strings.TrimSuffix(viper.GetString(config.StatusCheckURLKey), "/")
	if baseURL == "" {
		baseURL = strings.TrimSuffix(os.Getenv("STATUS_CHECK_URL"), "/")
	}
	if baseURL == "" {
		log.Warningf("session status HTTP check disabled: set status_check.url or STATUS_CHECK_URL to your server (e.g. https://api.example.com)")
		return
	}
	if u.token == "" {
		log.Warningf("session status HTTP check disabled: token is empty")
		return
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	ticker := time.NewTicker(statusCheckHTTPInterval)
	defer ticker.Stop()

	log.Infof("session status HTTP check started (interval %s)", statusCheckHTTPInterval)

	for {
		u.state.Mutex().Lock()
		st := u.state.GetStatus()
		sessionIDFromState := u.state.GetSessionID()
		u.state.Mutex().Unlock()

		sessionID := sessionIDFromState
		if sessionID == "" && u.currentSessionID != "" {
			sessionID = u.currentSessionID
		}
		if st != state.InRent {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			continue
		}
		if sessionID == "" {
			log.Warningf("session status HTTP check skipped: session_id empty (status=%s)", st)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			continue
		}

		url := baseURL + "/api/v1/status/rent/merchant/" + sessionID
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			continue
		}
		req.Header.Set("Authorization", "Bearer "+u.token)

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Warningf("session status HTTP check request failed: %v", err)
			if err := u.StopRequest(context.Background(), "session status check failed: "+err.Error()); err != nil {
				log.Errorf("failed to stop after HTTP check error: %s", err)
			}
			return
		}
		_ = resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			log.Infof("session [%s] no longer active (HTTP 404), stopping", sessionID)
			if err := u.StopRequest(context.Background(), "session no longer active (HTTP check)"); err != nil {
				log.Errorf("failed to stop after 404: %s", err)
			}
			return
		}
		if resp.StatusCode >= 500 {
			log.Warningf("session status HTTP check returned %d", resp.StatusCode)
		}

		select {
		case <-ctx.Done():
			log.Infof("session status HTTP check stopped")
			return
		case <-ticker.C:
		}
	}
}

// StartMerchantNodeCheckLoop запускает цикл проверки активных узлов (для статусов Ready и InRent).
func (u usecase) StartMerchantNodeCheckLoop(ctx context.Context) {
	u.merchantNodeHTTPCheck.Cancel()
	nodeHTTPCtx, nodeHTTPCancel := context.WithCancel(context.Background())
	u.merchantNodeHTTPCheck.SetCancelFn(nodeHTTPCancel)
	go func() { u.MerchantNodeCheckLoop(nodeHTTPCtx) }()
}

// CancelMerchantNodeCheckLoop останавливает цикл проверки узла.
func (u usecase) CancelMerchantNodeCheckLoop() {
	u.merchantNodeHTTPCheck.Cancel()
}

// MerchantNodeCheckLoop периодически проверяет по HTTP, активен ли ещё узел поставщика (не был ли остановлен через портал).
// Использует тот же baseURL что и StatusCheckLoop (userservice), endpoint /api/merchant/sessions.
// Если текущий session_id не найден в списке активных узлов — узел был остановлен, нужно отключиться.
func (u usecase) MerchantNodeCheckLoop(ctx context.Context) {
	baseURL := strings.TrimSuffix(viper.GetString(config.AuthServiceURLKey), "/")
	if baseURL == "" {
		baseURL = strings.TrimSuffix(os.Getenv("STATUS_CHECK_URL"), "/")
	}
	if baseURL == "" {
		log.Warningf("merchant node HTTP check disabled: set status_check.url or STATUS_CHECK_URL to your server (e.g. https://api.example.com)")
		return
	}
	if u.token == "" {
		log.Warningf("merchant node HTTP check disabled: token is empty")
		return
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	ticker := time.NewTicker(statusCheckHTTPInterval)
	defer ticker.Stop()

	log.Infof("merchant node HTTP check started (interval %s)", statusCheckHTTPInterval)

	for {
		u.state.Mutex().Lock()
		st := u.state.GetStatus()
		sessionIDFromState := u.state.GetSessionID()
		u.state.Mutex().Unlock()

		sessionID := sessionIDFromState
		if sessionID == "" && u.currentSessionID != "" {
			sessionID = u.currentSessionID
		}
		// Проверяем узел и при «Готов к приёму» (Ready), и при «В аренде» (InRent)
		if st != state.InRent && st != state.Ready {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			continue
		}
		if sessionID == "" {
			log.Warningf("merchant node HTTP check skipped: session_id empty (status=%s)", st)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			continue
		}

		url := baseURL + "/api/merchant/sessions"
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			continue
		}

		req.Header.Set("Authorization", "Bearer "+u.token)

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Warningf("merchant node HTTP check request failed: %v", err)
			if err := u.StopRequest(context.Background(), "merchant node check failed: "+err.Error()); err != nil {
				log.Errorf("failed to stop after HTTP check error: %s", err)
			}
			return
		}

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			_ = resp.Body.Close()
			log.Warningf("merchant node HTTP check returned %d (auth failed)", resp.StatusCode)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			log.Warningf("merchant node HTTP check returned %d", resp.StatusCode)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			continue
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			log.Warningf("merchant node HTTP check failed to read response: %v", err)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			continue
		}

		// Ответ: либо массив [...], либо объект с полем "data": [...]
		type sessionRow struct {
			SessionID string `json:"session_id"`
		}
		var sessions []sessionRow
		if err := json.Unmarshal(body, &sessions); err != nil {
			var wrapped struct {
				Data []sessionRow `json:"data"`
			}
			if err2 := json.Unmarshal(body, &wrapped); err2 != nil {
				log.Warningf("merchant node HTTP check failed to parse response: %v", err)
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
				}
				continue
			}
			sessions = wrapped.Data
		}
		if sessions == nil {
			sessions = []sessionRow{}
		}

		found := false
		for _, s := range sessions {
			if s.SessionID == sessionID {
				found = true
				break
			}
		}

		if !found {
			log.Infof("merchant node [%s] no longer in active nodes list, stopping (status=%s)", sessionID, st)
			if err := u.stopNodeFromPortal(context.Background()); err != nil {
				log.Errorf("failed to stop after node check: %s", err)
			}
			return
		}

		select {
		case <-ctx.Done():
			log.Infof("merchant node HTTP check stopped")
			return
		case <-ticker.C:
		}
	}
}
