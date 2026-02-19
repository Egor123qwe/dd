package history

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/domain/model"
	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/service"
	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/storage/repo"
)

type History interface {
	Rent(ctx *gin.Context)
	Lease(ctx *gin.Context)
	AdminRents(ctx *gin.Context)
}

type handler struct {
	service service.Sevice
}

func New(cfg *config.Config, log *slog.Logger) History {
	return handler{
		service: service.New(cfg, log),
	}
}

// @Summary History
// @Description This endpoint returns Rent History of User
// @Tags History
// @Accept json
// @Produce json
// @Success 200 {array} model.Rent
// @failure	404	{string} string "user has no history"
// @failure	500	{string} string "internal server error"
// @Router /api/v1/user/history/rent [get]
func (h handler) Rent(ctx *gin.Context) {
	req, ok := ctx.Get("UserID")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "userID is required"})
	}

	userID := req.(string)

	rents, err := h.service.History().Rent(userID)

	if err != nil {
		if errors.Is(err, repo.ErrNoHistory) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user has no history"})

			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})

		return
	}

	ctx.JSON(http.StatusOK, rents)
}

// @Summary History
// @Description This endpoint returns Session History of User
// @Tags History
// @Accept json
// @Produce json
// @Success 200 {array} model.Rent
// @failure	404	{string} string "user has no history"
// @failure	500	{string} string "internal server error"
// @Router /api/v1/user/history/lease [get]
func (h handler) Lease(ctx *gin.Context) {
	req, ok := ctx.Get("UserID")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "userID is required"})
	}

	userID := req.(string)

	rents, err := h.service.History().Session(userID)

	if err != nil {
		if errors.Is(err, repo.ErrNoHistory) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user has no history"})

			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})

		return
	}

	ctx.JSON(http.StatusOK, rents)
}

// AdminRents возвращает все завершённые аренды с client_id и merchant_id (для панели администратора).
func (h handler) AdminRents(ctx *gin.Context) {
	list, err := h.service.History().All()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if list == nil {
		list = []model.AdminRent{}
	}
	ctx.JSON(http.StatusOK, list)
}
