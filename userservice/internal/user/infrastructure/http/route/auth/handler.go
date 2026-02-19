package auth

import (
	refreshTokenUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/auth/refresh_token"
	registerUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/auth/register"
	signInUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/auth/sign_in"
	httpValidator "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/validator"
	"github.com/Interpuls/ifc2-service-farm/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type handler struct {
	log zerolog.Logger

	signInUC       signInUsecase.Usecase
	refreshTokenUC refreshTokenUsecase.Usecase
	registerUC     registerUsecase.Usecase

	validator httpValidator.Validator
}

func New(
	router *gin.RouterGroup,
	signInUC signInUsecase.Usecase,
	refreshTokenUC refreshTokenUsecase.Usecase,
	registerUC registerUsecase.Usecase,
	validator httpValidator.Validator,
) {
	h := &handler{
		log: logger.Named("AuthHandler"),

		signInUC:       signInUC,
		refreshTokenUC: refreshTokenUC,
		registerUC:     registerUC,
		validator:      validator,
	}

	router.POST("/signin", h.newSignInHandler())
	router.POST("/refresh", h.newRefreshTokenHandler())
	router.POST("/register", h.newRegisterHandler())
}
