package user

import (
	dto "github.com/Interpuls/ifc2-service-farm/internal/user/infrastructure/http/dto/user"
	balanceTopUpUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/balance_top_up"
	balanceWithdrawUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/balance_withdraw"
	createUserUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/create_user"
	deleteUserUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/delete_user"
	getBalanceUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/get_balance"
	getUserUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/get_user"
	getUserListUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/get_user_list"
	updateUserPasswordUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/update_user_password"
	updateUserProfileUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/update_user_profile"
	authMiddleware "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/middleware/auth"
	httpValidator "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/validator"
	"github.com/Interpuls/ifc2-service-farm/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type handler struct {
	log zerolog.Logger

	converter            *dto.Converter
	createUserUC         createUserUsecase.Usecase
	deleteUserUC         deleteUserUsecase.Usecase
	getUserUC            getUserUsecase.Usecase
	getUserListUC        getUserListUsecase.Usecase
	getBalanceUC         getBalanceUsecase.Usecase
	balanceTopUpUC       balanceTopUpUsecase.Usecase
	balanceWithdrawUC    balanceWithdrawUsecase.Usecase
	updateUserPasswordUC updateUserPasswordUsecase.Usecase
	updateUserProfileUC  updateUserProfileUsecase.Usecase
	middleware           authMiddleware.Middleware

	validator httpValidator.Validator
}

func New(
	router *gin.RouterGroup,
	createUserUC createUserUsecase.Usecase,
	deleteUserUC deleteUserUsecase.Usecase,
	getUserUC getUserUsecase.Usecase,
	getUserListUC getUserListUsecase.Usecase,
	getBalanceUC getBalanceUsecase.Usecase,
	balanceTopUpUC balanceTopUpUsecase.Usecase,
	balanceWithdrawUC balanceWithdrawUsecase.Usecase,
	updateUserPasswordUC updateUserPasswordUsecase.Usecase,
	updateUserProfileUC updateUserProfileUsecase.Usecase,
	middleware authMiddleware.Middleware,
	validator httpValidator.Validator,
) {
	converter := dto.NewConverter()

	h := &handler{
		log: logger.Named("UserHandler"),

		converter:            converter,
		createUserUC:         createUserUC,
		deleteUserUC:         deleteUserUC,
		getUserUC:            getUserUC,
		getUserListUC:        getUserListUC,
		getBalanceUC:         getBalanceUC,
		balanceTopUpUC:       balanceTopUpUC,
		balanceWithdrawUC:    balanceWithdrawUC,
		updateUserPasswordUC: updateUserPasswordUC,
		updateUserProfileUC:  updateUserProfileUC,
		middleware:           middleware,
		validator:            validator,
	}

	authMW := h.middleware.NewAuthMiddleware()

	router.POST("/", authMW, h.newCreateUserHandler())
	router.GET("/", authMW, h.newGetUserListHandler())
	router.GET("/:userId/balance", authMW, h.newGetBalanceHandler())
	router.POST("/:userId/balance/top-up", authMW, h.newBalanceTopUpHandler())
	router.POST("/:userId/balance/withdraw", authMW, h.newBalanceWithdrawHandler())
	router.GET("/:userId", authMW, h.newGetUserHandler())
	router.DELETE("/:userId", authMW, h.newDeleteUserHandler())
	router.PATCH("/:userId/password", authMW, h.newUpdateUserPasswordHandler())
	router.PATCH("/:userId/profile", authMW, h.newUpdateUserProfileHandler())
}
