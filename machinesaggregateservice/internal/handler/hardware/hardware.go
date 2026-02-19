package hardware

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
	log             slog.Logger
	hardwareService HardwareService
	tracer          trace.Tracer
}

type HardwareService interface {
	GPUList(ctx context.Context, log slog.Logger) ([]model.GPU, error)
	GPUByID(ctx context.Context, log slog.Logger, gpuID string) ([]model.GPU, error)
}

func New(log slog.Logger, hardwareService HardwareService, tracer trace.Tracer) Handler {
	h := Handler{
		log:             log,
		hardwareService: hardwareService,
		tracer:          tracer,
	}

	return h
}

func (h Handler) GPUByID(ctx *gin.Context) {
	c, span := h.tracer.Start(ctx.Request.Context(), "handler.hardware.GPUByID")
	defer span.End()

	gpuID := ctx.Param("gpu_id")

	gpus, err := h.hardwareService.GPUByID(c, h.log, gpuID)
	if err != nil {
		span.SetStatus(codes.Error, "failed to get GPUByID")
		span.RecordError(err)

		switch {
		case errors.Is(err, model.ErrGPUsNotFound):
			response.NewError(ctx, http.StatusNotFound, err.Error())

		default:
			response.NewError(ctx, http.StatusInternalServerError, err.Error())
		}

		return
	}

	span.SetAttributes(
		attribute.String("gpuID", gpuID),
		attribute.Int("gpus.count", len(gpus)),
	)

	ctx.JSON(http.StatusOK, gpus)
}

func (h Handler) GPUList(ctx *gin.Context) {
	c, span := h.tracer.Start(ctx.Request.Context(), "handler.hardware.GPUByID")
	defer span.End()

	gpus, err := h.hardwareService.GPUList(c, h.log)
	if err != nil {
		span.SetStatus(codes.Error, "failed to get GPUList")
		span.RecordError(err)

		switch {
		case errors.Is(err, model.ErrGPUsNotFound):
			response.NewError(ctx, http.StatusNotFound, err.Error())

		default:
			response.NewError(ctx, http.StatusInternalServerError, err.Error())
		}

		return
	}

	span.SetAttributes(
		attribute.Int("gpus.count", len(gpus)),
	)

	ctx.JSON(http.StatusOK, gpus)
}
