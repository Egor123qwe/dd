package category

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/handler/response"
	"gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/model"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrInvalidParams = errors.New("Invalid param value")
)

type Handler struct {
	categoryService CategoryService
	log             slog.Logger
	tracer          trace.Tracer
}

type CategoryService interface {
	GPUDictList(ctx context.Context, log slog.Logger, vramFrom, vramTo int) ([]model.GPUDict, error)
}

func New(log slog.Logger, categoryService CategoryService, tracer trace.Tracer) Handler {
	h := Handler{
		log:             log,
		categoryService: categoryService,
		tracer:          tracer,
	}

	return h
}

//	@Summary		List of gpu categoties
//	@Description	Create feeedback for rental.
//	@Accept			json
//	@Produce		json
//	@Param			q	query		string				false	"vram_from"
//	@Param			q	query		string				false	"vram_to"
//	@success		200	{array}		model.GPUDict		"Success receiving gpu categories"
//	@failure		404	{object}	response.HTTPError	"Not found"
//	@failure		500	{object}	response.HTTPError	"Something went wrong"
//	@Router			/exchange/category/gpu [get]
func (h Handler) GPUDict(ctx *gin.Context) {
	c, span := h.tracer.Start(ctx.Request.Context(), "handler.category.GPUDict")
	defer span.End()

	vramFromStr := ctx.DefaultQuery("vram_from", "0")
	vramToStr := ctx.DefaultQuery("vram_to", "999999999")

	vramFrom, err := strconv.Atoi(vramFromStr)
	if err != nil {
		span.SetStatus(codes.Error, "incorrect vramFrom")
		response.NewError(ctx, http.StatusBadRequest, ErrInvalidParams.Error())

		return
	}

	vramTo, err := strconv.Atoi(vramToStr)
	if err != nil {
		span.SetStatus(codes.Error, "incorrect vramTo")
		response.NewError(ctx, http.StatusBadRequest, ErrInvalidParams.Error())

		return
	}

	categories, err := h.categoryService.GPUDictList(c, h.log, vramFrom, vramTo)
	if err != nil {
		span.SetStatus(codes.Error, "failed to get GPUDictList")
		span.RecordError(err)

		switch {
		case errors.Is(err, model.ErrCatsNotFound):
			response.NewError(ctx, http.StatusNotFound, err.Error())

		default:
			response.NewError(ctx, http.StatusInternalServerError, err.Error())
		}

		return
	}

	span.SetAttributes(
		attribute.String("vramFrom", vramFromStr),
		attribute.String("vramTo", vramToStr),
		attribute.Int("categories.count", len(categories)),
	)

	ctx.JSON(http.StatusOK, categories)
}
