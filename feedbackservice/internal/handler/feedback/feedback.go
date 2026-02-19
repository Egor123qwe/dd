package feedback

import (
	"context"
	"errors"
	"net/http"

	"gitlab.roy9.ru/roy9/backend/core/feedbackservice/internal/domain/feedback"
	"gitlab.roy9.ru/roy9/backend/core/feedbackservice/internal/handler/response"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type Handler struct {
	feedbackService FeedbackService
}

type FeedbackService interface {
	Create(ctx context.Context, fb feedback.Feedback) (int64, error)
	CreateFeedbackLocal(ctx context.Context, fl feedback.FeedbackLocal) (int64, error)
	CreateFeedbackPartnership(ctx context.Context, fp feedback.FeedbackPartnership) (int64, error)
}

func New(feedbackSvc FeedbackService) *Handler {
	handler := &Handler{
		feedbackService: feedbackSvc,
	}

	return handler
}

// @Summary			Create feedback for rental endpoint
// @Description	Create feeedback for rental.
// @Accept			json
// @Produce			json
// @Param				request	body		feedback.Feedback	true	"feedback creating params"
// @success			200		{integer}	id	"Success create feedback"
// @failure			400		{object}	response.HTTPError	"Wrong body"
// @failure			500		{object}	response.HTTPError	"Somethink went wrong"
// @Router			/score/rental [post]	
func (h *Handler) Create(ctx *gin.Context) {
	tracer := otel.Tracer("feedbackservice")
	trCtx, span := tracer.Start(ctx.Request.Context(), "Create feedback for rental")
	defer span.End()

	var fb feedback.Feedback

	if err := ctx.ShouldBindJSON(&fb); err != nil {
		response.NewError(ctx, http.StatusBadRequest, feedback.ErrInvalidRequest.Error())

		return
	}

	span.SetAttributes(
		attribute.Int("feedback.score", fb.Score),
		attribute.String("feedback.rental_id", fb.RentID),
		attribute.String("feedback.comment", fb.Text),
	)

	id, err := h.feedbackService.Create(trCtx, fb)
	if err != nil {
		span.SetStatus(codes.Error, "failed to create feedback")
		span.RecordError(err)
		log.Error().Err(err).Msg("can not create feedback for rental")

		switch {
		case errors.Is(err, feedback.ErrFeedbackExists):
			response.NewError(ctx, http.StatusBadRequest, feedback.ErrFeedbackExists.Error())

		default:
			response.NewError(ctx, http.StatusInternalServerError, "something went wrong")
		}

		return
	}

	span.SetAttributes(attribute.Int64("feedback.id", id))
	ctx.JSON(http.StatusCreated, gin.H{
		"id": id,
	})
}

// @Summary			Create feedback for local endpoint
// @Description	Create feeedback for local.
// @Accept			json
// @Produce			json
// @Param				request	body		feedback.FeedbackLocal	true	"feedback creating params"
// @success			200		{integer}	id	"Success create feedback"
// @failure			400		{object}	response.HTTPError	"Wrong body"
// @failure			500		{object}	response.HTTPError	"Somethink went wrong"
// @Router			/request/local [post]
func (h *Handler) CreateFeedbackLocal(ctx *gin.Context) {
	tracer := otel.Tracer("feedbackservice")
	trCtx, span := tracer.Start(ctx.Request.Context(), "CreateFeedbackLocal")
	defer span.End()

	var fl feedback.FeedbackLocal

	if err := ctx.ShouldBindJSON(&fl); err != nil {
		log.Error().Err(err).Msg("can not parse request body")
		response.NewError(ctx, http.StatusBadRequest, feedback.ErrInvalidRequest.Error())

		return
	}

	id, ok := ctx.Get("userID")
	if !ok {
		log.Error().Msg("get user id")
		response.NewError(ctx, http.StatusInternalServerError, feedback.ErrInvalidRequest.Error())

		return
	}

	fl.UserID = id.(string)

	span.SetAttributes(
		attribute.String("feedback.local.user_id", fl.UserID),
		attribute.String("feedback.local.text", fl.Text),
		attribute.String("feedback.local.type", fl.Type),
	)

	feedbackLocalID, err := h.feedbackService.CreateFeedbackLocal(trCtx, fl)
	if err != nil {
		log.Error().Err(err).Msg("can not create feedback for local")
		span.SetStatus(codes.Error, "failed to create feedback for local")
		span.RecordError(err)

		switch {
		default:
			response.NewError(ctx, http.StatusInternalServerError, "something went wrong")
		}

		return
	}

	span.SetAttributes(attribute.Int64("feedback.id.local", feedbackLocalID))
	ctx.JSON(http.StatusCreated, gin.H{
		"id": id,
	})
}

// @Summary			Create feedback for partnership endpoint
// @Description	Create feeedback for partnership.
// @Accept			json
// @Produce			json
// @Param				request	body		feedback.FeedbackPartnership	true	"feedback creating params"
// @success			200		{integer}	id	"Success create feedback"
// @failure			400		{object}	response.HTTPError	"Wrong body"
// @failure			500		{object}	response.HTTPError	"Somethink went wrong"
// @Router			/request/partnership [post]
func (h *Handler) CreateFeedbackPartnership(ctx *gin.Context) {
	tracer := otel.Tracer("feedbackservice")
	trCtx, span := tracer.Start(ctx.Request.Context(), "CreateFeedbackLocal")
	defer span.End()

	var fp feedback.FeedbackPartnership

	if err := ctx.ShouldBindJSON(&fp); err != nil {
		log.Error().Err(err).Msg("can not parse request body")
		response.NewError(ctx, http.StatusBadRequest, feedback.ErrInvalidRequest.Error())

		return
	}

	id, ok := ctx.Get("userID")
	if !ok {
		log.Error().Msg("get user id")
		response.NewError(ctx, http.StatusInternalServerError, feedback.ErrInvalidRequest.Error())

		return
	}

	fp.UserID = id.(string)

	span.SetAttributes(
		attribute.String("feedback.partnership.user_id", fp.UserID),
		attribute.String("feedback.partnership.phone_number", fp.PhoneNum),
		attribute.String("feedback.partnership.email", fp.Email),
		attribute.String("feedback.partnership.comment", fp.Comment),
		attribute.String("feedback.partnership.company_name", fp.CompanyName),
		attribute.String("feedback.partnership.contact_name", fp.ContactName),
	)

	feedbackPartnershipID, err := h.feedbackService.CreateFeedbackPartnership(trCtx, fp)
	if err != nil {
		log.Error().Err(err).Msg("can not create feedback for partnership")
		span.SetStatus(codes.Error, "failed to create feedback for patnership")
		span.RecordError(err)

		switch {
		default:
			response.NewError(ctx, http.StatusInternalServerError, "something went wrong")
		}

		return
	}

	span.SetAttributes(attribute.Int64("feedback.id.partnership", feedbackPartnershipID))
	ctx.JSON(http.StatusCreated, gin.H{
		"id": id,
	})
}
