package handler

import (
	"net/http"

	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/service"
)

type handler struct {
	router *http.ServeMux
	srv    service.Service
}

func New(srv service.Service) http.Handler {
	h := &handler{
		router: http.NewServeMux(),
		srv:    srv,
	}

	h.initWSRoutes()

	return h.router
}
