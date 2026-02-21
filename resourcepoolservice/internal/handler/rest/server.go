package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/repository/merchant"

	"github.com/rs/zerolog/log"
)

// SessionListerWithDetails возвращает список сессий с деталями и список темплейтов.
type SessionListerWithDetails interface {
	ListReadySessionsWithDetails(ctx context.Context) ([]merchant.ReadySessionDetails, error)
	ListAllTemplateIDs(ctx context.Context) ([]string, error)
	GetSessionDetailsByID(ctx context.Context, sessionID string) (*merchant.ReadySessionDetails, error)
}

// MerchantSessionLister расширяет интерфейс для портала поставщика (мои узлы, отключить).
type MerchantSessionLister interface {
	SessionListerWithDetails
	ListSessionsByUserID(ctx context.Context, userID string) ([]merchant.MerchantSessionRow, error)
	Stop(ctx context.Context, sessionID, deletionReason string) error
}

// RentStopper останавливает аренду по session_id узла (вызов sessionhandlerservice). Может быть nil.
type RentStopper interface {
	StopRentByMerchantSessionID(ctx context.Context, sessionID string) error
}

type Server struct {
	port        string
	list        MerchantSessionLister
	rentStopper RentStopper
}

func New(port string, list MerchantSessionLister, rentStopper RentStopper) *Server {
	return &Server{port: port, list: list, rentStopper: rentStopper}
}

type merchantResp struct {
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id"`
	UserID    string                 `json:"user_id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	NodeName  string                 `json:"node_name,omitempty"`
	Details   *merchantDetailsResp    `json:"details,omitempty"`
}

type merchantDetailsResp struct {
	TotalPrice    float64                   `json:"total_price"`
	TotalRAM      int64                     `json:"total_ram"`
	AvailableRAM  int64                     `json:"available_ram"`
	PriceRAM      float64                   `json:"price_ram"`
	LoadSpeed     float64                   `json:"load_speed"`
	UploadSpeed   float64                   `json:"upload_speed"`
	Ping          int64                     `json:"ping"`
	PriceInternet float64                   `json:"price_internet"`
	GPUs          []merchant.SessionGPU     `json:"gpus"`
	CPUs          []merchant.SessionCPU     `json:"cpus"`
	Storages      []merchant.SessionStorage `json:"storages"`
	Templates     []string                  `json:"templates"`
}

func (s *Server) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/client/rent/merchants", s.handleListMerchants)
	mux.HandleFunc("/api/client/rent/merchant/", s.handleGetMerchantBySessionID)
	mux.HandleFunc("/api/client/rent/templates", s.handleListTemplates)
	mux.HandleFunc("/api/merchant/sessions", s.handleMerchantSessions)
	mux.HandleFunc("/api/merchant/session/", s.handleMerchantSessionStop)

	server := &http.Server{Addr: ":" + s.port, Handler: mux}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	log.Info().Str("port", s.port).Msg("REST server listening")
	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) handleListMerchants(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	sessions, err := s.list.ListReadySessionsWithDetails(ctx)
	if err != nil {
		log.Error().Err(err).Msg("list ready sessions with details")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	out := make([]merchantResp, 0, len(sessions))
	for _, sess := range sessions {
		details := &merchantDetailsResp{
			TotalPrice:    sess.TotalPrice,
			TotalRAM:      sess.TotalRAM,
			AvailableRAM:  sess.AvailableRAM,
			PriceRAM:      sess.PriceRAM,
			LoadSpeed:     sess.LoadSpeed,
			UploadSpeed:   sess.UploadSpeed,
			Ping:          sess.Ping,
			PriceInternet: sess.PriceInternet,
			GPUs:          sess.GPUs,
			CPUs:          sess.CPUs,
			Storages:      sess.Storages,
			Templates:     sess.Templates,
		}
		out = append(out, merchantResp{
			ID:        sess.ID,
			SessionID: sess.ID,
			UserID:    sess.UserID,
			Name:      "",
			NodeName:  sess.NodeName,
			Details:   details,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(out); err != nil {
		log.Error().Err(err).Msg("encode merchants response")
	}
}

func (s *Server) handleGetMerchantBySessionID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	sessionID := strings.TrimPrefix(r.URL.Path, "/api/client/rent/merchant/")
	sessionID = strings.Trim(sessionID, "/")
	if sessionID == "" {
		http.Error(w, "session_id required", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	sess, err := s.list.GetSessionDetailsByID(ctx, sessionID)
	if err != nil {
		log.Error().Err(err).Msg("get session details by id")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if sess == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	details := &merchantDetailsResp{
		TotalPrice:    sess.TotalPrice,
		TotalRAM:      sess.TotalRAM,
		AvailableRAM:  sess.AvailableRAM,
		PriceRAM:      sess.PriceRAM,
		LoadSpeed:     sess.LoadSpeed,
		UploadSpeed:   sess.UploadSpeed,
		Ping:          sess.Ping,
		PriceInternet: sess.PriceInternet,
		GPUs:          sess.GPUs,
		CPUs:          sess.CPUs,
		Storages:      sess.Storages,
		Templates:     sess.Templates,
	}
	out := merchantResp{
		ID:        sess.ID,
		SessionID: sess.ID,
		UserID:    sess.UserID,
		NodeName:  sess.NodeName,
		Details:   details,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(out); err != nil {
		log.Error().Err(err).Msg("encode merchant response")
	}
}

func (s *Server) handleListTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	ids, err := s.list.ListAllTemplateIDs(ctx)
	if err != nil {
		log.Error().Err(err).Msg("list all template ids")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if ids == nil {
		ids = []string{}
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ids); err != nil {
		log.Error().Err(err).Msg("encode templates response")
	}
}

// handleMerchantSessions — GET /api/merchant/sessions, заголовок X-User-ID.
func (s *Server) handleMerchantSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := strings.TrimSpace(r.Header.Get("X-User-ID"))
	if userID == "" {
		http.Error(w, "X-User-ID required", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	rows, err := s.list.ListSessionsByUserID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("list sessions by user id")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	out := make([]struct {
		SessionID  string  `json:"session_id"`
		NodeName   string  `json:"node_name"`
		Status     string  `json:"status"`
		TotalPrice float64 `json:"total_price"`
	}, 0, len(rows))
	for _, row := range rows {
		out = append(out, struct {
			SessionID  string  `json:"session_id"`
			NodeName   string  `json:"node_name"`
			Status     string  `json:"status"`
			TotalPrice float64 `json:"total_price"`
		}{SessionID: row.ID, NodeName: row.NodeName, Status: row.Status, TotalPrice: row.TotalPrice})
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(out); err != nil {
		log.Error().Err(err).Msg("encode merchant sessions response")
	}
}

// handleMerchantSessionStop — POST /api/merchant/session/:id/stop, заголовок X-User-ID.
func (s *Server) handleMerchantSessionStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := strings.TrimSpace(r.Header.Get("X-User-ID"))
	if userID == "" {
		http.Error(w, "X-User-ID required", http.StatusBadRequest)
		return
	}
	sessionID := strings.TrimPrefix(r.URL.Path, "/api/merchant/session/")
	sessionID = strings.TrimSuffix(sessionID, "/stop")
	sessionID = strings.Trim(sessionID, "/")
	if sessionID == "" {
		http.Error(w, "session_id required", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	sess, err := s.list.GetSessionDetailsByID(ctx, sessionID)
	if err != nil || sess == nil {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}
	if sess.UserID != userID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err := s.list.Stop(ctx, sessionID, "stopped from portal"); err != nil {
		log.Error().Err(err).Msg("stop session")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if s.rentStopper != nil {
		if err := s.rentStopper.StopRentByMerchantSessionID(ctx, sessionID); err != nil {
			log.Error().Err(err).Str("session_id", sessionID).Msg("stop rent by merchant session_id (node already marked stopped)")
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
