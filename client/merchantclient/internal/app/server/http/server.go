package http

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/config"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/usecase/rent"
)

var log = logging.MustGetLogger("http")

//go:embed static
var staticFS embed.FS

// Backend is the web UI backend: connect with token, then start/stop rent.
type Backend interface {
	Rent() (rent.Usecase, bool)
	Connect(ctx context.Context, token string) error
	Disconnect()
	GetHardware(ctx context.Context) (map[string]interface{}, error)
}

type Server struct {
	backend     Backend
	port        int
	authBaseURL string
	authClient  *http.Client
}

func New(backend Backend) Server {
	port := viper.GetInt(config.HttpPortKey)
	if port == 0 {
		port = 8080
	}
	authBaseURL := strings.TrimSuffix(viper.GetString(config.AuthServiceURLKey), "/")
	return Server{
		backend:     backend,
		port:        port,
		authBaseURL: authBaseURL,
		authClient:  &http.Client{},
	}
}

func (s Server) Serve(ctx context.Context) error {
	static, _ := fs.Sub(staticFS, "static")

	mux := http.NewServeMux()
	mux.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("merchantclient ok"))
	})
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/connect", s.handleConnect)
	mux.HandleFunc("/api/disconnect", s.handleDisconnect)
	mux.HandleFunc("/api/rent/start", s.handleRentStart)
	mux.HandleFunc("/api/rent/stop", s.handleRentStop)
	mux.HandleFunc("/api/hardware", s.handleHardware)
	mux.HandleFunc("/api/auth/signin", s.handleAuthSignIn)
	mux.HandleFunc("/api/auth/refresh", s.handleAuthRefresh)
	mux.Handle("/", http.FileServer(http.FS(static)))

	addr := ":" + fmtPort(s.port)
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Infof("HTTP server listening on %s", addr)

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	rentUC, connected := s.backend.Rent()
	if !connected {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"connected":   false,
			"rent_status": nil,
			"rent_time":   nil,
		})
		return
	}
	status := rentUC.GetStatus().String()
	resp := map[string]interface{}{
		"connected":   true,
		"rent_status": status,
		"rent_time":   nil,
		"total_price": nil,
	}
	if status == "InRent" {
		if startedAt := rentUC.GetRentStartedAt(); startedAt != nil {
			d := time.Since(*startedAt)
			h := int(d.Hours())
			m := int(d.Minutes()) % 60
			sec := int(d.Seconds()) % 60
			resp["rent_time"] = fmt.Sprintf("%02d:%02d:%02d", h, m, sec)
		}
	}
	// Получаем total_price из state (сохранено при init)
	if price := rentUC.GetTotalPrice(); price > 0 {
		resp["total_price"] = price
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s Server) handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := s.backend.Connect(context.Background(), body.Token); err != nil {
		log.Warningf("connect failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s Server) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.backend.Disconnect()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s Server) handleAuthSignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.authBaseURL == "" {
		http.Error(w, "auth service not configured (set auth.service_url)", http.StatusServiceUnavailable)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	req, _ := http.NewRequestWithContext(r.Context(), http.MethodPost, s.authBaseURL+"/api/auth/signin", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.authClient.Do(req)
	if err != nil {
		log.Warningf("auth signin proxy: %v", err)
		http.Error(w, "auth service unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (s Server) handleAuthRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.authBaseURL == "" {
		http.Error(w, "auth service not configured (set auth.service_url)", http.StatusServiceUnavailable)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	req, _ := http.NewRequestWithContext(r.Context(), http.MethodPost, s.authBaseURL+"/api/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.authClient.Do(req)
	if err != nil {
		log.Warningf("auth refresh proxy: %v", err)
		http.Error(w, "auth service unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (s Server) handleRentStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rentUC, connected := s.backend.Rent()
	if !connected {
		http.Error(w, "not connected", http.StatusServiceUnavailable)
		return
	}
	var body struct {
		CheatMode bool   `json:"cheat_mode"`
		NodeName  string `json:"node_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := rentUC.Init(r.Context(), body.CheatMode, body.NodeName); err != nil {
		log.Warningf("rent start failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s Server) handleRentStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rentUC, connected := s.backend.Rent()
	if !connected {
		http.Error(w, "not connected", http.StatusServiceUnavailable)
		return
	}
	if err := rentUC.StopRequest(r.Context()); err != nil {
		log.Warningf("rent stop failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s Server) handleHardware(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	hw, err := s.backend.GetHardware(r.Context())
	if err != nil {
		log.Warningf("hardware fetch failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(hw)
}

func fmtPort(p int) string {
	if p == 0 {
		return "8080"
	}
	return strconv.Itoa(p)
}
