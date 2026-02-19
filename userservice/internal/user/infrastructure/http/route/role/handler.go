package role

import (
	dto "github.com/Interpuls/ifc2-service-farm/internal/user/infrastructure/http/dto/role"
	createRoleUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/role/create_role"
	deleteRoleUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/role/delete_role"
	getRoleListUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/role/get_role_list"
	updateRoleUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/role/update_role"
	authMiddleware "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/middleware/auth"
	httpValidator "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/validator"
	"github.com/gin-gonic/gin"
)

type handler struct {
	converter     dto.Converter
	createRoleUC  createRoleUsecase.Usecase
	deleteRoleUC  deleteRoleUsecase.Usecase
	getRoleListUC getRoleListUsecase.Usecase
	updateRoleUC  updateRoleUsecase.Usecase
	middleware    authMiddleware.Middleware

	validator httpValidator.Validator
}

func New(
	router *gin.RouterGroup,
	createRoleUC createRoleUsecase.Usecase,
	deleteRoleUC deleteRoleUsecase.Usecase,
	getRoleListUC getRoleListUsecase.Usecase,
	updateRoleUC updateRoleUsecase.Usecase,
	middleware authMiddleware.Middleware,
	validator httpValidator.Validator,
) {
	h := &handler{
		converter:     dto.NewConverter(),
		createRoleUC:  createRoleUC,
		deleteRoleUC:  deleteRoleUC,
		getRoleListUC: getRoleListUC,
		updateRoleUC:  updateRoleUC,
		middleware:    middleware,
		validator:     validator,
	}

	authMW := h.middleware.NewAuthMiddleware()

	router.POST("/", authMW, h.newCreateRoleHandler())
	router.GET("/", authMW, h.newGetRoleListHandler())
	router.PATCH("/:roleId", authMW, h.newUpdateRoleHandler())
	router.DELETE("/:roleId", authMW, h.newDeleteRoleHandler())
}
