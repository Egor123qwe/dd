package session

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/handler/response"
	"gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/model"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	log            slog.Logger
	sessionService SessionService
	tracer         trace.Tracer
}

type SessionService interface {
	ReceiveByID(ctx context.Context, log slog.Logger, sessionID string) (model.SessionResponse, error)
	ListByTariffID(ctx context.Context, log slog.Logger, tariffID string) (model.SessionList, error)
	ListByGPUID(ctx context.Context, log slog.Logger, gpuDictID string) (model.SessionList, error)
}

func New(log slog.Logger, sessionService SessionService, tracer trace.Tracer) Handler {
	h := Handler{
		log:            log,
		sessionService: sessionService,
		tracer:         tracer,
	}

	return h
}

// @Summary			Receive session endpoint
// @Description	This endpoint returns session by ID.
// @Accept			json
// @Produce			json
// @Param				session_id	path		string					true	"sessionID"
// @success			200					{object}	model.SessionResponse	"Success receiving session"
// @failure			404					{object}	response.HTTPError		"Not found"
// @failure			500					{object}	response.HTTPError		"Something went wrong"
// @Router			/session/{session_id} [get]
func (h Handler) ReceiveByID(ctx *gin.Context) {
	c, span := h.tracer.Start(ctx.Request.Context(), "handler.session.ReceiveByID")
	defer span.End()

	sessionID := ctx.Param("session_id")

	session, err := h.sessionService.ReceiveByID(c, h.log, sessionID)
	if err != nil {
		span.SetStatus(codes.Error, "failed to receive ByID")
		span.RecordError(err)

		switch {
		case errors.Is(err, model.ErrSessionsNotFound):
			response.NewError(ctx, http.StatusNotFound, err.Error())

		default:
			response.NewError(ctx, http.StatusInternalServerError, err.Error())
		}

		return
	}

	span.SetAttributes(
		attribute.String("sessionID", sessionID),
	)

	ctx.JSON(http.StatusOK, session)
}

// @Summary			Receive list sessions by tariffID endpoint
// @Description	This endpoint returns sessions by ID of tariff.
// @Accept			json
// @Produce			json
// @Param				tariff_id	path		string					true	"tariffID"
// @success			200					{object}	model.SessionList			"Success receiving sessions"
// @failure			404					{object}	response.HTTPError		"Not found"
// @failure			500					{object}	response.HTTPError		"Something went wrong"
// @Router			/exchange/category/tariff/{tariff_id}/session [get]
func (h Handler) ListByTariffID(ctx *gin.Context) {
	c, span := h.tracer.Start(ctx.Request.Context(), "handler.session.ListByTariffID")
	defer span.End()

	tariffID := ctx.Param("tariff_id")

	sessions, err := h.sessionService.ListByTariffID(c, h.log, tariffID)
	if err != nil {
		span.SetStatus(codes.Error, "failed to get ListByTariffID")
		span.RecordError(err)

		switch {
		case errors.Is(err, model.ErrSessionsNotFound):
			response.NewError(ctx, http.StatusNotFound, err.Error())

		default:
			response.NewError(ctx, http.StatusInternalServerError, err.Error())
		}

		return
	}

	span.SetAttributes(
		attribute.String("tariffID", tariffID),
		attribute.Int("sessions.count", len(sessions.Sessions)),
	)

	ctx.JSON(http.StatusOK, sessions)
}

// @Summary			Receive list sessions by gpuIDlist endpoint
// @Description	This endpoint returns sessions by ID of gpulist.
// @Accept			json
// @Produce			json
// @Param				gpu_id	path		string					true	"gpuID"
// @success			200					{object}	model.SessionList			"Success receiving sessions"
// @failure			404					{object}	response.HTTPError		"Not found"
// @failure			500					{object}	response.HTTPError		"Something went wrong"
// @Router			/exchange/category/gpu/{gpu_id}/session [get]
func (h Handler) ListByGPUID(ctx *gin.Context) {
	c, span := h.tracer.Start(ctx.Request.Context(), "handler.session.ListByGPUID")
	defer span.End()

	gpuID := ctx.Param("gpu_id")

	sessions, err := h.sessionService.ListByGPUID(c, h.log, gpuID)
	if err != nil {
		span.SetStatus(codes.Error, "failed to get ListByGPUID")
		span.RecordError(err)

		switch {
		case errors.Is(err, model.ErrSessionsNotFound):
			response.NewError(ctx, http.StatusNotFound, err.Error())

		default:
			response.NewError(ctx, http.StatusInternalServerError, err.Error())
		}

		return
	}

	span.SetAttributes(
		attribute.String("gpuID", gpuID),
		attribute.Int("sessions.count", len(sessions.Sessions)),
	)

	ctx.JSON(http.StatusOK, sessions)
}
