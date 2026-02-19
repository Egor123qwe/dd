package user

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Interpuls/ifc2-service-farm/config"
	"github.com/Interpuls/ifc2-service-farm/internal/user/infrastructure/http"
	balanceRepo "github.com/Interpuls/ifc2-service-farm/internal/user/infrastructure/mysql/balance"
	roleRepo "github.com/Interpuls/ifc2-service-farm/internal/user/infrastructure/mysql/role"
	userRepo "github.com/Interpuls/ifc2-service-farm/internal/user/infrastructure/mysql/user"
	"github.com/Interpuls/ifc2-service-farm/internal/user/infrastructure/postgres"
	"github.com/Interpuls/ifc2-service-farm/pkg/jwt"
	"github.com/gin-gonic/gin"

	// Token UCs
	refreshTokenUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/auth/refresh_token"
	registerUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/auth/register"
	signInUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/auth/sign_in"

	// Role UCs
	createRoleUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/role/create_role"
	deleteRoleUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/role/delete_role"
	getRoleListUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/role/get_role_list"
	updateRoleUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/role/update_role"

	// User UCs
	balanceTopUpUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/balance_top_up"
	balanceWithdrawUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/balance_withdraw"
	createUserUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/create_user"
	settleRentUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/settle_rent"
	deleteUserUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/delete_user"
	getBalanceUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/get_balance"
	getUserUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/get_user"
	getUserListUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/get_user_list"
	getUsernameByIDUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/get_username_by_id"
	updateUserPasswordUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/update_user_password"
	updateUserProfileUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/update_user_profile"
)

type Module struct {
	ucs ucs
}

type ucs struct {
	// Token UCs
	SignInUC       signInUsecase.Usecase
	RefreshTokenUC refreshTokenUsecase.Usecase
	RegisterUC     registerUsecase.Usecase

	// Role UCs
	CreateRoleUC  createRoleUsecase.Usecase
	DeleteRoleUC  deleteRoleUsecase.Usecase
	GetRoleListUC getRoleListUsecase.Usecase
	UpdateRoleUC  updateRoleUsecase.Usecase

	// User UCs
	CreateUserUC         createUserUsecase.Usecase
	DeleteUserUC         deleteUserUsecase.Usecase
	GetUserUC            getUserUsecase.Usecase
	GetUserListUC        getUserListUsecase.Usecase
	GetUsernameByIDUC    getUsernameByIDUsecase.Usecase
	GetBalanceUC         getBalanceUsecase.Usecase
	BalanceTopUpUC       balanceTopUpUsecase.Usecase
	BalanceWithdrawUC    balanceWithdrawUsecase.Usecase
	SettleRentUC         settleRentUsecase.Usecase
	UpdateUserPasswordUC updateUserPasswordUsecase.Usecase
	UpdateUserProfileUC  updateUserProfileUsecase.Usecase
}

func Init(
	cfg config.UserConfig,
	jwtService jwt.JWT,
) (Module, error) {

	pgStoreConfig := postgres.Config{
		URL:            cfg.Postgres.URL,
		MigrationsDir:  cfg.Postgres.MigrationsDir,
		TestSeedersDir: cfg.Postgres.TestSeedersDir,
		RunTestSeeders: cfg.Postgres.RunTestSeeders,
	}

	pgStore, err := postgres.New(pgStoreConfig)
	if err != nil {
		return Module{}, err
	}

	if err := pgStore.UpdateSchema(); err != nil {
		return Module{}, fmt.Errorf("failed to update schema: %w", err)
	}

	if err := pgStore.RunTestSeeders(); err != nil {
		return Module{}, fmt.Errorf("failed to run test seeders: %w", err)
	}

	userRepository := userRepo.New(pgStore.DB())
	roleRepository := roleRepo.New(pgStore.DB())
	balanceRepository := balanceRepo.New(pgStore.DB())

	signInUsecaseConfig := signInUsecase.Config{
		AtExp: cfg.Token.AtExp,
		RtExp: cfg.Token.RtExp,
	}

	refreshTokenUsecaseConfig := refreshTokenUsecase.Config{
		AtExp: cfg.Token.AtExp,
		RtExp: cfg.Token.RtExp,
	}

	registerUsecaseConfig := registerUsecase.Config{
		DefaultRoleID: cfg.Registration.DefaultRoleID,
		AtExp:         cfg.Token.AtExp,
		RtExp:         cfg.Token.RtExp,
	}

	ucs := ucs{
		// Token UCs
		SignInUC:       signInUsecase.New(userRepository, roleRepository, jwtService, signInUsecaseConfig),
		RefreshTokenUC: refreshTokenUsecase.New(userRepository, roleRepository, jwtService, refreshTokenUsecaseConfig),
		RegisterUC:     registerUsecase.New(userRepository, roleRepository, jwtService, registerUsecaseConfig),

		// Role UCs
		CreateRoleUC:  createRoleUsecase.New(roleRepository),
		DeleteRoleUC:  deleteRoleUsecase.New(roleRepository, userRepository),
		GetRoleListUC: getRoleListUsecase.New(roleRepository, userRepository),
		UpdateRoleUC:  updateRoleUsecase.New(roleRepository, userRepository),

		// User UCs
		CreateUserUC:         createUserUsecase.New(userRepository, roleRepository),
		DeleteUserUC:         deleteUserUsecase.New(userRepository, roleRepository),
		GetUserUC:            getUserUsecase.New(userRepository, roleRepository),
		GetUserListUC:        getUserListUsecase.New(userRepository, roleRepository),
		GetUsernameByIDUC:    getUsernameByIDUsecase.New(userRepository),
		GetBalanceUC:         getBalanceUsecase.New(balanceRepository),
		BalanceTopUpUC:       balanceTopUpUsecase.New(balanceRepository),
		BalanceWithdrawUC:    balanceWithdrawUsecase.New(balanceRepository),
		SettleRentUC:         settleRentUsecase.New(balanceRepository),
		UpdateUserPasswordUC: updateUserPasswordUsecase.New(userRepository, roleRepository),
		UpdateUserProfileUC:  updateUserProfileUsecase.New(userRepository, roleRepository),
	}

	module := Module{
		ucs: ucs,
	}

	return module, nil
}

func (m Module) AssignHttpHandler(router *gin.RouterGroup, jwt jwt.JWT) {
	getUsernameByID := func(ctx context.Context, userID string) (string, error) {
		id, err := strconv.Atoi(userID)
		if err != nil {
			return "", err
		}
		return m.ucs.GetUsernameByIDUC.GetUsernameByID(ctx, id)
	}
	http.Register(
		router, jwt,
		getUsernameByID,
		// Token UCs
		m.ucs.SignInUC,
		m.ucs.RefreshTokenUC,
		m.ucs.RegisterUC,

		// Role UCs
		m.ucs.CreateRoleUC,
		m.ucs.DeleteRoleUC,
		m.ucs.GetRoleListUC,
		m.ucs.UpdateRoleUC,

		// User UCs
		m.ucs.CreateUserUC,
		m.ucs.DeleteUserUC,
		m.ucs.GetUserUC,
		m.ucs.GetUserListUC,
		m.ucs.GetBalanceUC,
		m.ucs.BalanceTopUpUC,
		m.ucs.BalanceWithdrawUC,
		m.ucs.SettleRentUC,
		m.ucs.UpdateUserPasswordUC,
		m.ucs.UpdateUserProfileUC,
	)
}
