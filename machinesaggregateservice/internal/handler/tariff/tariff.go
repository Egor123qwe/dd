package tariff

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
	tariffSerivce TariffService
	log           slog.Logger
	tracer        trace.Tracer
}

type TariffService interface {
	List(ctx context.Context, log slog.Logger) ([]model.Tariff, error)
}

func New(log slog.Logger, tariffSerivce TariffService, tracer trace.Tracer) Handler {
	h := Handler{
		log:           log,
		tariffSerivce: tariffSerivce,
		tracer:        tracer,
	}

	return h
}

//	@Summary		List of tariffs endpoint
//	@Description	This endpoint returns information about tariffs and information about available sessions.
//	@Accept			json
//	@Produce		json
//	@success		200	{array}		model.Tariff		"Success receiving tariffs"
//	@failure		404	{object}	response.HTTPError	"Not found"
//	@failure		500	{object}	response.HTTPError	"Something went wrong"
//	@Router			/exchange/category/tariff [get]
func (h Handler) List(ctx *gin.Context) {
	c, span := h.tracer.Start(ctx.Request.Context(), "handler.tariff.List")
	defer span.End()

	tariffs, err := h.tariffSerivce.List(c, h.log)
	if err != nil {
		span.SetStatus(codes.Error, "failed to get list of tariffs")
		span.RecordError(err)

		switch {
		case errors.Is(err, model.ErrCatsNotFound):
			response.NewError(ctx, http.StatusNotFound, model.ErrCatsNotFound.Error())

		default:
			response.NewError(ctx, http.StatusInternalServerError, err.Error())
		}

		return
	}

	span.SetAttributes(
		attribute.Int("tariffs.count", len(tariffs)),
	)

	ctx.JSON(http.StatusOK, tariffs)
}
