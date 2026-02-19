package server

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/domain/model"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/handler/model/message"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/service"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/service/status"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/storage"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/storage/db/repo"
)

var (
	ErrBadRequest     = errors.New("bad request")
	ErrInternalServer = errors.New("internal server")
)

type Handler interface {
	HandleClientStatus(ctx *gin.Context)
	HandleMerchantStatus(ctx *gin.Context)
	HandleClientStatusBySession(ctx *gin.Context)
	HandleMerchantSessions(ctx *gin.Context)
	HandleMerchantSession(ctx *gin.Context)
}

type handler struct {
	service service.Service
	log     slog.Logger
}

func New(cfg config.Config, log slog.Logger, storage storage.Storage) Handler {

	service := service.New(cfg, log, storage)

	return handler{
		service: service,
		log:     log,
	}
}

// @Summary Status
// @Description This endpoint returns Rent Status of User
// @Tags Status
// @Accept json
// @Produce json
// @Success 200 {array} model.Client
// @failure	404	{string} string "user has no rent"
// @failure	500	{string} string "internal server error"
// @Router /api/v1/status/rent/client [get]
func (h handler) HandleClientStatus(ctx *gin.Context) {
	var (
		clientMsg message.ClientMessage
	)

	req, ok := ctx.Get("UserID")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "userID is required"})

		return
	}
	userID := req.(string)
	h.log.Info("HandleClientStatus", "userID", userID)

	clientMsg.UserID = userID

	clientRent, err := h.service.Status().GetClientRent(ctx, clientMsg)

	if err != nil {
		if errors.Is(err, status.ErrClienttNoRents) {
			ctx.JSON(http.StatusOK, make([]model.Client, 0))
			return
		}
		h.log.Error("can't get client rent:", err.Error(), nil)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, clientRent)

}

// @Summary Status
// @Description This endpoint returns Rent Statuses of User
// @Tags Status
// @Accept json
// @Produce json
// @Param session_id body string true "Session ID"
// @Success 200 {object} model.Merchant
// @failure	404	{string} string "user has no rent"
// @failure	500	{string} string "internal server error"
// @Router /api/v1/status/rent/merchant/{session_id} [get]
func (h handler) HandleMerchantStatus(ctx *gin.Context) {
	var merchantMsg message.MerchantMessage

	req, ok := ctx.Get("UserID")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "userID is required"})

		return
	}

	userID := req.(string)

	merchantMsg.SessionID = ctx.Param("session_id")

	if merchantMsg.SessionID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})

		return
	}

	merchantMsg.UserID = userID

	merchantRent, err := h.service.Status().GetMerchantRent(ctx, merchantMsg)

	if err != nil {
		if errors.Is(err, status.ErrMerchantNoSessions) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Merchant has no sessions"})

			return
		}

		h.log.Error("can't get merchant rent:", err.Error(), nil)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer})

		return
	}

	ctx.JSON(http.StatusOK, merchantRent)
}

// @Summary Status
// @Description This endpoint returns Rent Status of User
// @Tags Status
// @Accept json
// @Produce json
// @Param session_id body string true "Session ID"
// @Success 200 {object} model.Client
// @failure	404	{string} string "user has no rent"
// @failure	500	{string} string "internal server error"
// @Router /api/v1/status/rent/client/{session_id} [get]
func (h handler) HandleClientStatusBySession(ctx *gin.Context) {
	var clientMsg message.ClientMessage

	req, ok := ctx.Get("UserID")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "userID is required"})

		return
	}
	userID := req.(string)

	clientMsg.UserID = userID

	clientMsg.SessionID = ctx.Param("session_id")

	if clientMsg.SessionID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})

		return
	}

	clientRent, err := h.service.Status().GetClientRentBySession(ctx, clientMsg)

	if err != nil {
		if errors.Is(err, status.ErrClienttNoRents) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Client has no Rent for this session_id"})

			return
		}

		h.log.Error("can't get client rent:", err.Error(), nil)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer})

		return
	}

	ctx.JSON(http.StatusOK, clientRent)

}

// @Summary Status
// @Description This endpoint returns session list
// @Tags Status
// @Accept json
// @Produce json
// @Success 200 {array} string "session_id list"
// @failure	404	{string} string "user has no sessions"
// @failure	500	{string} string "internal server error"
// @Router /api/v1/status/session [get]
func (h handler) HandleMerchantSessions(ctx *gin.Context) {
	req, ok := ctx.Get("UserID")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "userID is required"})

		return
	}
	userID := req.(string)

	sessions, err := h.service.Status().MerchantSessions(ctx, userID)

	if err != nil {
		if errors.Is(err, repo.ErrSessionDoesNotExist) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Merchant has no sessions"})

			return
		}

		h.log.Error("can't get merchant sessions:", err.Error(), nil)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer})

		return
	}

	ctx.JSON(http.StatusOK, sessions)
}

// @Summary Status
// @Description This endpoint returns session list
// @Tags Status
// @Accept json
// @Produce json
// @Success 200 {object} model.Session "merchant session"
// @failure	404	{string} string "user has no sessions"
// @failure	500	{string} string "internal server error"
// @Router /api/v1/status/session/{session_id} [get]
func (h handler) HandleMerchantSession(ctx *gin.Context) {
	req, ok := ctx.Get("UserID")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "userID is required"})

		return
	}
	userID := req.(string)

	sessionID := ctx.Param("session_id")

	if sessionID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})

		return 
	}

	session, err := h.service.Status().DetailMerchantSession(ctx, sessionID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrSessionDoesNotExist) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Merchant has no sessions"})

			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer})

		return
	}

	ctx.JSON(http.StatusOK, session)
}
