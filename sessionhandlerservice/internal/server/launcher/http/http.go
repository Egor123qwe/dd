package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/rent"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/session"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/server/launcher"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/service"
	rentRepo "gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/db/psql/repo/rent"
)

var logHTTP = logging.MustGetLogger("http")

type server struct {
	srv  service.Service
	port string
}

func New(srv service.Service) (launcher.Server, error) {
	port := viper.GetString("server.http.port")
	if port == "" {
		port = "8096"
	}
	return &server{srv: srv, port: port}, nil
}

func (s *server) Serve(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/client/rent/templates", s.handleListTemplates)
	mux.HandleFunc("/api/client/rent/stop", s.handleClientRentStop)
	mux.HandleFunc("/api/client/rent/settings", s.handleClientRentSettings)
	mux.HandleFunc("/api/admin/templates", s.handleAdminTemplatesCreate)
	mux.HandleFunc("/api/admin/templates/", s.handleAdminTemplatesByID)
	mux.HandleFunc("/internal/merchant/session/", s.handleInternalStopByMerchantSessionID)

	httpServer := &http.Server{Addr: ":" + s.port, Handler: mux}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()
	logHTTP.Infof("HTTP server listening on :%s", s.port)
	return httpServer.ListenAndServe()
}

type portResp struct {
	Auth  bool   `json:"auth"`
	Port  int    `json:"port"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

type envResp struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

type templateInfoResp struct {
	TemplateID       string     `json:"template_id"`
	Title            string     `json:"title"`
	Type             string     `json:"type"`
	Description      string     `json:"description"`
	ShortDescription string     `json:"short_description"`
	Version          string     `json:"version"`
	ContainerImage   string     `json:"container_image_name"`
	ImageTag         string     `json:"container_image_tag"`
	Ports            []portResp `json:"ports"`
	Envs             []envResp  `json:"envs"`
	Volumes          []string   `json:"volumes"`
	UseGPU           bool       `json:"use_gpu"`
}

func templateToResp(t rent.Template) templateInfoResp {
	ports := make([]portResp, 0, len(t.Ports))
	for _, p := range t.Ports {
		ports = append(ports, portResp{Auth: p.Auth, Port: p.Port, Type: p.Type, Title: p.Title})
	}
	envs := make([]envResp, 0, len(t.Envs))
	for _, e := range t.Envs {
		envs = append(envs, envResp{Key: e.Key, Value: e.Value, Type: e.Type})
	}
	return templateInfoResp{
		TemplateID:       t.ID,
		Title:            t.Title,
		Type:             t.Type,
		Description:      t.Description,
		ShortDescription: t.ShortDescription,
		Version:          t.Version,
		ContainerImage:   t.ImageName,
		ImageTag:         t.ImageTag,
		Ports:            ports,
		Envs:             envs,
		Volumes:          t.Volumes,
		UseGPU:           t.UseGPU,
	}
}

func (s *server) handleListTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	list, err := s.srv.ListTemplates(ctx)
	if err != nil {
		logHTTP.Errorf("list templates: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	out := make([]templateInfoResp, 0, len(list))
	for _, t := range list {
		out = append(out, templateToResp(t))
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(out); err != nil {
		logHTTP.Errorf("encode templates response: %v", err)
	}
}

type adminPortPayload struct {
	Auth  bool   `json:"auth"`
	Port  int    `json:"port"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

type adminEnvPayload struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

type adminTemplatePayload struct {
	Title            string               `json:"title"`
	Description      string               `json:"description"`
	ShortDescription string               `json:"short_description"`
	ContainerImage   string               `json:"container_image_name"`
	ImageTag         string               `json:"container_image_tag"`
	Ports            []adminPortPayload   `json:"ports"`
	Envs             []adminEnvPayload     `json:"envs"`
	Volumes          []string             `json:"volumes"`
	UseGPU           bool                 `json:"use_gpu"`
}

func adminPayloadToPorts(p []adminPortPayload) []rent.Port {
	out := make([]rent.Port, 0, len(p))
	for _, x := range p {
		out = append(out, rent.Port{Auth: x.Auth, Port: x.Port, Type: x.Type, Title: x.Title})
	}
	return out
}

func adminPayloadToEnvs(p []adminEnvPayload) []rent.Env {
	out := make([]rent.Env, 0, len(p))
	for _, x := range p {
		out = append(out, rent.Env{Key: x.Key, Value: x.Value, Type: x.Type})
	}
	return out
}

func (s *server) handleAdminTemplatesCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	var body adminTemplatePayload
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ports := adminPayloadToPorts(body.Ports)
	envs := adminPayloadToEnvs(body.Envs)
	volumes := body.Volumes
	if volumes == nil {
		volumes = []string{}
	}
	t, err := s.srv.CreateTemplate(ctx, body.Title, body.Description, body.ShortDescription, body.ContainerImage, body.ImageTag, body.UseGPU, ports, envs, volumes)
	if err != nil {
		logHTTP.Errorf("create template: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(templateToResp(t)); err != nil {
		logHTTP.Errorf("encode template response: %v", err)
	}
}

func (s *server) handleAdminTemplatesByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/admin/templates/")
	id = strings.Trim(id, "/")
	if id == "" {
		http.Error(w, "template id required", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	switch r.Method {
	case http.MethodGet:
		t, err := s.srv.GetTemplate(ctx, id)
		if err != nil {
			http.Error(w, "template not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(templateToResp(t)); err != nil {
			logHTTP.Errorf("encode template: %v", err)
		}
		return
	case http.MethodPatch, http.MethodPut:
		var body adminTemplatePayload
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		ports := adminPayloadToPorts(body.Ports)
		envs := adminPayloadToEnvs(body.Envs)
		volumes := body.Volumes
		if volumes == nil {
			volumes = []string{}
		}
		if err := s.srv.UpdateTemplate(ctx, id, body.Title, body.Description, body.ShortDescription, body.ContainerImage, body.ImageTag, body.UseGPU, ports, envs, volumes); err != nil {
			logHTTP.Errorf("update template: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		t, err := s.srv.GetTemplate(ctx, id)
		if err != nil {
			_ = json.NewEncoder(w).Encode(map[string]string{"template_id": id})
			return
		}
		_ = json.NewEncoder(w).Encode(templateToResp(t))
		return
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

type stopRentReq struct {
	RequestID string `json:"request_id"`
	Reason    string `json:"reason"`
}

func (s *server) handleClientRentStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := strings.TrimSpace(r.Header.Get("X-User-ID"))
	if userID == "" {
		http.Error(w, "X-User-ID header required", http.StatusUnauthorized)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	var body stopRentReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	body.RequestID = strings.TrimSpace(body.RequestID)
	if body.RequestID == "" {
		http.Error(w, "request_id required", http.StatusBadRequest)
		return
	}

	rentClientID, err := s.srv.GetRentClientID(ctx, body.RequestID)
	if err != nil {
		logHTTP.Errorf("get rent client id: %v", err)
		http.Error(w, "rent not found", http.StatusNotFound)
		return
	}
	if rentClientID != userID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	_, err = s.srv.Session().Stop(ctx, session.StopReq{
		RequestID: body.RequestID,
		Reason:    strings.TrimSpace(body.Reason),
	})
	if err != nil {
		logHTTP.Errorf("client rent stop: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}

// handleInternalStopByMerchantSessionID — POST /internal/merchant/session/:sessionId/stop.
// Вызывается resourcepoolservice при отключении узла с портала: останавливает аренду по session_id узла.
func (s *server) handleInternalStopByMerchantSessionID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	sessionID := strings.TrimPrefix(r.URL.Path, "/internal/merchant/session/")
	sessionID = strings.TrimSuffix(sessionID, "/stop")
	sessionID = strings.Trim(sessionID, "/")
	if sessionID == "" {
		http.Error(w, "session_id required", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	requestID, err := s.srv.Session().GetMerchantRent(ctx, sessionID)
	if err != nil {
		if errors.Is(err, rentRepo.ErrRentNotFound) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true}`))
			return
		}
		logHTTP.Errorf("get merchant rent by session_id: %v", err)
		http.Error(w, "failed to get rent", http.StatusInternalServerError)
		return
	}

	_, err = s.srv.Session().Stop(ctx, session.StopReq{
		RequestID: requestID,
		Reason:    "node disconnected from portal",
	})
	if err != nil {
		logHTTP.Errorf("internal stop by merchant session_id: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func (s *server) handleClientRentSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := strings.TrimSpace(r.Header.Get("X-User-ID"))
	if userID == "" {
		http.Error(w, "X-User-ID header required", http.StatusUnauthorized)
		return
	}
	requestID := strings.TrimSpace(r.URL.Query().Get("request_id"))
	if requestID == "" {
		http.Error(w, "request_id query required", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	rentClientID, err := s.srv.GetRentClientID(ctx, requestID)
	if err != nil {
		logHTTP.Errorf("get rent client id: %v", err)
		http.Error(w, "rent not found", http.StatusNotFound)
		return
	}
	if rentClientID != userID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	settings, err := s.srv.Rent().GetRentSettings(ctx, requestID)
	if err != nil {
		logHTTP.Errorf("get rent settings: %v", err)
		http.Error(w, "failed to get settings", http.StatusInternalServerError)
		return
	}

	type endpointResp struct {
		Title string `json:"title,omitempty"`
		Link  string `json:"link,omitempty"`
	}
	type pikoResp struct {
		Endpoints []endpointResp `json:"endpoints,omitempty"`
	}
	type networkResp struct {
		Piko *pikoResp `json:"piko,omitempty"`
	}
	type authResp struct {
		Login    string `json:"login,omitempty"`
		Password string `json:"password,omitempty"`
	}
	type templateResp struct {
		Login          string    `json:"login,omitempty"`
		Password       string    `json:"password,omitempty"`
		Authentication *authResp `json:"authentication,omitempty"`
	}
	out := struct {
		Template *templateResp `json:"template,omitempty"`
		Network  *networkResp `json:"network,omitempty"`
	}{}
	out.Template = &templateResp{
		Login:    settings.Template.Authentication.Login,
		Password: settings.Template.Authentication.Password,
		Authentication: &authResp{
			Login:    settings.Template.Authentication.Login,
			Password: settings.Template.Authentication.Password,
		},
	}
	if settings.Network.Piko != nil {
		eps := make([]endpointResp, 0, len(settings.Network.Piko.Endpoints))
		for _, e := range settings.Network.Piko.Endpoints {
			eps = append(eps, endpointResp{Title: e.Title, Link: e.Link})
		}
		out.Network = &networkResp{Piko: &pikoResp{Endpoints: eps}}
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(out); err != nil {
		logHTTP.Errorf("encode settings response: %v", err)
	}
}
