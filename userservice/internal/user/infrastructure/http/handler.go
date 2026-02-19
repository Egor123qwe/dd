package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	authHandler "github.com/Interpuls/ifc2-service-farm/internal/user/infrastructure/http/route/auth"
	permissionsHandler "github.com/Interpuls/ifc2-service-farm/internal/user/infrastructure/http/route/permissions"
	roleHandler "github.com/Interpuls/ifc2-service-farm/internal/user/infrastructure/http/route/role"
	userHandler "github.com/Interpuls/ifc2-service-farm/internal/user/infrastructure/http/route/user"
	"github.com/Interpuls/ifc2-service-farm/internal/user/infrastructure/http/validation"
	authMiddleware "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/middleware/auth"
	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	"github.com/Interpuls/ifc2-service-farm/pkg/jwt"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"

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
	updateUserPasswordUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/update_user_password"
	updateUserProfileUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/update_user_profile"

	"github.com/gin-gonic/gin"
)

// GetUsernameByID возвращает username по user_id (для обогащения списка мерчантов). Может быть nil.
type GetUsernameByID func(ctx context.Context, userID string) (string, error)

func Register(
	router *gin.RouterGroup,
	jwt jwt.JWT,
	getUsernameByID GetUsernameByID,

	// Token UCs
	signInUC signInUsecase.Usecase,
	refreshTokenUC refreshTokenUsecase.Usecase,
	registerUC registerUsecase.Usecase,

	// Role UCs
	createRoleUC createRoleUsecase.Usecase,
	deleteRoleUC deleteRoleUsecase.Usecase,
	getRoleListUC getRoleListUsecase.Usecase,
	updateRoleUC updateRoleUsecase.Usecase,

	// User UCs
	createUserUC createUserUsecase.Usecase,
	deleteUserUC deleteUserUsecase.Usecase,
	getUserUC getUserUsecase.Usecase,
	getUserListUC getUserListUsecase.Usecase,
	getBalanceUC getBalanceUsecase.Usecase,
	balanceTopUpUC balanceTopUpUsecase.Usecase,
	balanceWithdrawUC balanceWithdrawUsecase.Usecase,
	settleRentUC settleRentUsecase.Usecase,
	updateUserPasswordUC updateUserPasswordUsecase.Usecase,
	updateUserProfileUC updateUserProfileUsecase.Usecase,
) {
	validator := validation.New()
	authMW := authMiddleware.New(jwt, nil)

	// Token routes
	authGroup := router.Group("/auth")
	authHandler.New(
		authGroup,
		signInUC,
		refreshTokenUC,
		registerUC,
		validator,
	)

	// User routes
	userGroup := router.Group("/users")
	userHandler.New(
		userGroup,
		createUserUC,
		deleteUserUC,
		getUserUC,
		getUserListUC,
		getBalanceUC,
		balanceTopUpUC,
		balanceWithdrawUC,
		updateUserPasswordUC,
		updateUserProfileUC,
		authMW, validator,
	)

	// Role routes
	roleGroup := router.Group("/roles")
	roleHandler.New(
		roleGroup,
		createRoleUC,
		deleteRoleUC,
		getRoleListUC,
		updateRoleUC,
		authMW, validator,
	)

	// PermissionsNames routes
	permissionsGroup := router.Group("/permissions")
	permissionsHandler.New(permissionsGroup, authMW)

	// Machine token validate: for connection coordinator (Client-Authorization / merchant client).
	// GET /api/v1/machine/token/validate with Authorization: Bearer <access_token>
	// Returns 200 {"user_id": "<id>"} or 403 on invalid/expired token.
	machineGroup := router.Group("/v1/machine")
	machineGroup.GET("/token/validate", validateMachineToken(jwt))
	// POST /api/v1/auth/validate — для healthcheckservice (FetchUserMiddleware): проверка токена, ответ {"user_id": "<id>"}.
	authValidateGroup := router.Group("/v1/auth")
	authValidateGroup.POST("/validate", validateAuthForServices(jwt))

	// Client rent: проксирование списка мерчантов в resourcepoolservice + обогащение именем мерчанта.
	// GET /api/client/rent/merchants -> ResourcePoolURL/api/client/rent/merchants, затем подстановка name по user_id
	// GET /api/client/rent/active -> HealthCheckURL/api/v1/status/rent/client (первый элемент массива как активная аренда)
	// GET /api/client/rent/history -> HistoryServiceURL/api/v1/user/history/rent
	// GET /api/client/rent/templates -> SessionHandlerHTTPURL/api/client/rent/templates (отдельный запрос, таблица templates_template_info)
	// POST /api/client/rent/:rentId/stop -> SessionHandlerHTTPURL/api/client/rent/stop (body request_id, reason; header X-User-ID)
	clientRentGroup := router.Group("/client/rent")
	clientRentGroup.GET("/active", proxyClientRentActive(getUsernameByID))
	clientRentGroup.GET("/history", proxyClientRentHistory())
	clientRentGroup.GET("/merchant/:sessionId", proxyClientRentMerchantDetails())
	clientRentGroup.GET("/:rentId/settings", proxyClientRentSettings(jwt))
	clientRentGroup.GET("/merchants", proxyClientRentMerchants(getUsernameByID))
	clientRentGroup.GET("/templates", proxyClientRentTemplates())
	clientRentGroup.POST("/:rentId/stop", proxyClientRentStop(jwt))

	// Merchant (поставщик): мои узлы, история сдачи в аренду, отключение узла.
	// GET /api/merchant/sessions -> RESOURCE_POOL_URL/api/merchant/sessions (X-User-ID из JWT)
	// GET /api/merchant/history/lease -> HISTORY_SERVICE_URL/api/v1/user/history/lease (Authorization)
	// GET /api/merchant/session/:sessionId/active-rent -> HEALTH_CHECK_URL/api/v1/status/rent/merchant/:sessionId (Authorization)
	// POST /api/merchant/session/:sessionId/stop -> RESOURCE_POOL_URL/api/merchant/session/:sessionId/stop (X-User-ID из JWT)
	merchantGroup := router.Group("/merchant")
	merchantGroup.GET("/sessions", proxyMerchantSessions(jwt))
	merchantGroup.GET("/history/lease", proxyMerchantHistoryLease())
	merchantGroup.GET("/session/:sessionId/active-rent", proxyMerchantSessionActiveRent())
	merchantGroup.POST("/session/:sessionId/stop", proxyMerchantSessionStop(jwt))

	// Внутренний API для sessionhandlerservice: списание с клиента и начисление продавцу при завершении аренды.
	internalBalanceGroup := router.Group("/internal/balance")
	internalBalanceGroup.POST("/settle-rent", handleInternalSettleRent(settleRentUC))

	// Админ: все аренды и шаблоны (только при пермишене SystemSettingsWrite).
	adminGroup := router.Group("/admin")
	adminGroup.Use(authMW.NewAuthMiddleware())
	adminGroup.GET("/history/rents", proxyAdminHistoryRents(getUsernameByID))
	adminGroup.POST("/templates", proxyAdminTemplatesCreate())
	adminGroup.GET("/templates/:templateId", proxyAdminTemplatesGet())
	adminGroup.PATCH("/templates/:templateId", proxyAdminTemplatesUpdate())
}

type settleRentReq struct {
	ClientUserID   int     `json:"client_user_id"`
	MerchantUserID int     `json:"merchant_user_id"`
	Amount         float64 `json:"amount"`
}

func handleInternalSettleRent(uc settleRentUsecase.Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
			return
		}
		var body settleRentReq
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if body.Amount < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be non-negative"})
			return
		}
		err := uc.SettleRent(c.Request.Context(), body.ClientUserID, body.MerchantUserID, body.Amount, settleRentUsecase.DefaultMerchantRate)
		if err != nil {
			if errors.Is(err, errs.ErrInsufficientBalance) {
				c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient balance"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

func validateMachineToken(jwtSrv jwt.JWT) gin.HandlerFunc {
	return validateAuthForServices(jwtSrv)
}

// validateAuthForServices проверяет Bearer-токен и возвращает 200 {"user_id": "<id>"}; иначе 403. Используется для machine/token/validate и auth/validate (healthcheckservice).
func validateAuthForServices(jwtSrv jwt.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "missing Authorization header"})
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid Authorization format"})
			return
		}
		claims, err := jwtSrv.Validate(parts[1])
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid or expired token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": strconv.Itoa(claims.UserID)})
	}
}

// merchantItem — элемент списка от resourcepool (для обогащения полем name).
type merchantItem struct {
	ID        string          `json:"id"`
	SessionID string          `json:"session_id"`
	UserID    string          `json:"user_id,omitempty"`
	Name      string          `json:"name,omitempty"`
	NodeName  string          `json:"node_name,omitempty"`
	Details   json.RawMessage `json:"details,omitempty"`
}

// proxyClientRentMerchants проксирует GET /api/client/rent/merchants в resourcepoolservice и подставляет name (username мерчанта).
func proxyClientRentMerchants(getUsernameByID GetUsernameByID) gin.HandlerFunc {
	baseURL := os.Getenv("RESOURCE_POOL_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8091"
	}
	client := &http.Client{Timeout: 15 * time.Second}
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
			return
		}
		url := baseURL + "/api/client/rent/merchants"
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusOK, []interface{}{})
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusOK, []interface{}{})
			return
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response"})
			return
		}
		var list []merchantItem
		if err := json.Unmarshal(body, &list); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response"})
			return
		}
		ctx := c.Request.Context()
		for i := range list {
			if getUsernameByID != nil && list[i].UserID != "" {
				name, _ := getUsernameByID(ctx, list[i].UserID)
				list[i].Name = name
			}
		}
		c.Header("Content-Type", "application/json")
		c.Status(http.StatusOK)
		_ = json.NewEncoder(c.Writer).Encode(list)
	}
}

// proxyClientRentMerchantDetails проксирует GET /api/client/rent/merchant/:sessionId в resourcepoolservice (характеристики мерчанта по session_id).
func proxyClientRentMerchantDetails() gin.HandlerFunc {
	baseURL := os.Getenv("RESOURCE_POOL_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8091"
	}
	client := &http.Client{Timeout: 15 * time.Second}
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
			return
		}
		sessionID := c.Param("sessionId")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId required"})
			return
		}
		url := strings.TrimSuffix(baseURL, "/") + "/api/client/rent/merchant/" + sessionID
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
			return
		}
		if auth := c.GetHeader("Authorization"); auth != "" {
			req.Header.Set("Authorization", auth)
		}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "resource pool unavailable", "detail": err.Error()})
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response"})
			return
		}
		if resp.StatusCode == http.StatusNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "merchant not found"})
			return
		}
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
	}
}

// proxyClientRentActive проксирует GET /api/client/rent/active в healthcheckservice; возвращает 200 и первый элемент массива (id = request_id, merchant_name по merchant_user_id) или null.
func proxyClientRentActive(getUsernameByID GetUsernameByID) gin.HandlerFunc {
	baseURL := os.Getenv("HEALTH_CHECK_URL")
	if baseURL == "" {
		baseURL = "http://localhost:19000"
	}
	client := &http.Client{Timeout: 15 * time.Second}
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
			return
		}
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required (Bearer token)"})
			return
		}
		url := strings.TrimSuffix(baseURL, "/") + "/api/v1/status/rent/client"
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
			return
		}
		req.Header.Set("Authorization", auth)
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "healthcheck service unavailable", "detail": err.Error()})
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		if resp.StatusCode != http.StatusOK {
			c.Data(http.StatusOK, "application/json", []byte("null"))
			return
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response"})
			return
		}
		var list []map[string]interface{}
		if err := json.Unmarshal(body, &list); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response"})
			return
		}
		if len(list) == 0 {
			c.Data(http.StatusOK, "application/json", []byte("null"))
			return
		}
		first := list[0]
		first["id"] = first["request_id"]
		if getUsernameByID != nil {
			if mid, ok := first["merchant_user_id"].(string); ok && mid != "" {
				if name, _ := getUsernameByID(c.Request.Context(), mid); name != "" {
					first["merchant_name"] = name
				}
			}
		}
		c.Header("Content-Type", "application/json")
		c.Status(http.StatusOK)
		_ = json.NewEncoder(c.Writer).Encode(first)
	}
}

// proxyClientRentSettings проксирует GET /api/client/rent/:rentId/settings в sessionhandlerservice (GET /api/client/rent/settings?request_id=... с X-User-ID).
func proxyClientRentSettings(jwtSrv jwt.JWT) gin.HandlerFunc {
	baseURL := os.Getenv("SESSION_HANDLER_HTTP_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8096"
	}
	client := &http.Client{Timeout: 15 * time.Second}
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
			return
		}
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required (Bearer token)"})
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization format"})
			return
		}
		claims, err := jwtSrv.Validate(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		rentId := strings.TrimSpace(c.Param("rentId"))
		if rentId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "rentId required"})
			return
		}
		url := strings.TrimSuffix(baseURL, "/") + "/api/client/rent/settings?request_id=" + rentId
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
			return
		}
		req.Header.Set("X-User-ID", strconv.Itoa(claims.UserID))
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "session handler unavailable", "detail": err.Error()})
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response"})
			return
		}
		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden {
			c.JSON(resp.StatusCode, gin.H{"error": "rent not found or access denied"})
			return
		}
		if resp.StatusCode != http.StatusOK {
			c.JSON(resp.StatusCode, gin.H{"error": string(body)})
			return
		}
		c.Header("Content-Type", "application/json")
		c.Status(http.StatusOK)
		_, _ = c.Writer.Write(body)
	}
}

// proxyClientRentHistory проксирует GET /api/client/rent/history в historyservice (история аренд клиента).
func proxyClientRentHistory() gin.HandlerFunc {
	baseURL := os.Getenv("HISTORY_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8899"
	}
	client := &http.Client{Timeout: 15 * time.Second}
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
			return
		}
		url := strings.TrimSuffix(baseURL, "/") + "/api/v1/user/history/rent"
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
			return
		}
		if auth := c.GetHeader("Authorization"); auth != "" {
			req.Header.Set("Authorization", auth)
		}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusOK, []interface{}{})
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			c.Header("Content-Type", "application/json")
			c.Status(http.StatusOK)
			_, _ = c.Writer.Write([]byte("[]"))
			return
		}
		if resp.StatusCode != http.StatusOK {
			c.JSON(resp.StatusCode, gin.H{"error": "history service error"})
			return
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response"})
			return
		}
		c.Header("Content-Type", "application/json")
		c.Status(http.StatusOK)
		_, _ = c.Writer.Write(body)
	}
}

// proxyClientRentStop проксирует POST /api/client/rent/:rentId/stop в sessionhandlerservice (POST /api/client/rent/stop с X-User-ID и body request_id, reason).
func proxyClientRentStop(jwtSrv jwt.JWT) gin.HandlerFunc {
	baseURL := os.Getenv("SESSION_HANDLER_HTTP_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8096"
	}
	client := &http.Client{Timeout: 15 * time.Second}
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
			return
		}
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required (Bearer token)"})
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization format"})
			return
		}
		claims, err := jwtSrv.Validate(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		rentId := strings.TrimSpace(c.Param("rentId"))
		if rentId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "rentId required"})
			return
		}
		var body struct {
			Reason string `json:"reason"`
		}
		_ = c.ShouldBindJSON(&body)
		payload := map[string]string{"request_id": rentId, "reason": body.Reason}
		jsonBody, _ := json.Marshal(payload)
		url := strings.TrimSuffix(baseURL, "/") + "/api/client/rent/stop"
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, url, strings.NewReader(string(jsonBody)))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", strconv.Itoa(claims.UserID))
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "session handler service unavailable", "detail": err.Error()})
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "rent not found"})
			return
		}
		if resp.StatusCode == http.StatusForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			c.JSON(resp.StatusCode, gin.H{"error": string(bodyBytes)})
			return
		}
		c.Status(http.StatusOK)
		_, _ = io.Copy(c.Writer, resp.Body)
	}
}

// proxyClientRentTemplates проксирует GET /api/client/rent/templates в sessionhandlerservice (таблица templates_template_info).
func proxyClientRentTemplates() gin.HandlerFunc {
	baseURL := os.Getenv("SESSION_HANDLER_HTTP_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8096"
	}
	client := &http.Client{Timeout: 15 * time.Second}
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
			return
		}
		url := strings.TrimSuffix(baseURL, "/") + "/api/client/rent/templates"
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusOK, []interface{}{})
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			c.JSON(resp.StatusCode, gin.H{"error": "templates service error"})
			return
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response"})
			return
		}
		c.Header("Content-Type", "application/json")
		c.Status(http.StatusOK)
		_, _ = c.Writer.Write(body)
	}
}

// proxyMerchantSessions проксирует GET /api/merchant/sessions в resourcepool (X-User-ID из JWT).
func proxyMerchantSessions(jwtSrv jwt.JWT) gin.HandlerFunc {
	baseURL := os.Getenv("RESOURCE_POOL_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8091"
	}
	client := &http.Client{Timeout: 15 * time.Second}
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
			return
		}
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required (Bearer token)"})
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization format"})
			return
		}
		claims, err := jwtSrv.Validate(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		url := strings.TrimSuffix(baseURL, "/") + "/api/merchant/sessions"
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
			return
		}
		req.Header.Set("X-User-ID", strconv.Itoa(claims.UserID))
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "resource pool unavailable", "detail": err.Error()})
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response"})
			return
		}
		if resp.StatusCode != http.StatusOK {
			c.JSON(resp.StatusCode, gin.H{"error": string(body)})
			return
		}
		c.Header("Content-Type", "application/json")
		c.Status(http.StatusOK)
		_, _ = c.Writer.Write(body)
	}
}

// proxyMerchantSessionActiveRent проксирует GET /api/merchant/session/:sessionId/active-rent в healthcheckservice (активная аренда по сессии: created_at и т.д.).
func proxyMerchantSessionActiveRent() gin.HandlerFunc {
	baseURL := os.Getenv("HEALTH_CHECK_URL")
	if baseURL == "" {
		baseURL = "http://localhost:19000"
	}
	client := &http.Client{Timeout: 15 * time.Second}
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
			return
		}
		sessionId := c.Param("sessionId")
		if sessionId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId required"})
			return
		}
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required (Bearer token)"})
			return
		}
		url := strings.TrimSuffix(baseURL, "/") + "/api/v1/status/rent/merchant/" + sessionId
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
			return
		}
		req.Header.Set("Authorization", auth)
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "healthcheck service unavailable", "detail": err.Error()})
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response"})
			return
		}
		if resp.StatusCode == http.StatusNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "no active rent for this session"})
			return
		}
		if resp.StatusCode != http.StatusOK {
			c.JSON(resp.StatusCode, gin.H{"error": string(body)})
			return
		}
		c.Header("Content-Type", "application/json")
		c.Status(http.StatusOK)
		_, _ = c.Writer.Write(body)
	}
}

// proxyMerchantHistoryLease проксирует GET /api/merchant/history/lease в historyservice (история сдачи в аренду).
func proxyMerchantHistoryLease() gin.HandlerFunc {
	baseURL := os.Getenv("HISTORY_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8899"
	}
	client := &http.Client{Timeout: 15 * time.Second}
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
			return
		}
		url := strings.TrimSuffix(baseURL, "/") + "/api/v1/user/history/lease"
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
			return
		}
		if auth := c.GetHeader("Authorization"); auth != "" {
			req.Header.Set("Authorization", auth)
		}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusOK, []interface{}{})
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			c.Header("Content-Type", "application/json")
			c.Status(http.StatusOK)
			_, _ = c.Writer.Write([]byte("[]"))
			return
		}
		if resp.StatusCode != http.StatusOK {
			c.JSON(resp.StatusCode, gin.H{"error": "history service error"})
			return
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response"})
			return
		}
		c.Header("Content-Type", "application/json")
		c.Status(http.StatusOK)
		_, _ = c.Writer.Write(body)
	}
}

// proxyMerchantSessionStop проксирует POST /api/merchant/session/:sessionId/stop в resourcepool (X-User-ID из JWT).
func proxyMerchantSessionStop(jwtSrv jwt.JWT) gin.HandlerFunc {
	baseURL := os.Getenv("RESOURCE_POOL_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8091"
	}
	client := &http.Client{Timeout: 15 * time.Second}
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
			return
		}
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required (Bearer token)"})
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization format"})
			return
		}
		claims, err := jwtSrv.Validate(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		sessionId := strings.TrimSpace(c.Param("sessionId"))
		if sessionId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId required"})
			return
		}
		url := strings.TrimSuffix(baseURL, "/") + "/api/merchant/session/" + sessionId + "/stop"
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, url, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
			return
		}
		req.Header.Set("X-User-ID", strconv.Itoa(claims.UserID))
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "resource pool unavailable", "detail": err.Error()})
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		if resp.StatusCode == http.StatusForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		if resp.StatusCode != http.StatusOK {
			c.JSON(resp.StatusCode, gin.H{"error": string(body)})
			return
		}
		c.Header("Content-Type", "application/json")
		c.Status(http.StatusOK)
		_, _ = c.Writer.Write(body)
	}
}

// adminRentResponse — элемент ответа GET /api/admin/history/rents (с именами покупателя и поставщика).
type adminRentResponse struct {
	ID            string  `json:"id"`
	ClientID      string  `json:"client_id"`
	MerchantID    string  `json:"merchant_id"`
	ClientName    string  `json:"client_name"`
	MerchantName  string  `json:"merchant_name"`
	Cost          float32 `json:"cost"`
	Duration      int     `json:"duration"`
	StartedAt     *string `json:"started_at,omitempty"`
	EndedAt       *string `json:"ended_at,omitempty"`
	TemplateID    string  `json:"template_id"`
	TemplateTitle string  `json:"template_title"`
}

// proxyAdminHistoryRents проксирует GET /api/admin/history/rents в historyservice и обогащает именами пользователей.
func proxyAdminHistoryRents(getUsernameByID GetUsernameByID) gin.HandlerFunc {
	baseURL := os.Getenv("HISTORY_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8899"
	}
	client := &http.Client{Timeout: 20 * time.Second}
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
			return
		}
		permit := authMiddleware.GetPermitFromContext(c)
		if permit == nil || !permit.HasPermission(permission.SystemSettingsWrite) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: SystemSettingsWrite required"})
			return
		}
		url := strings.TrimSuffix(baseURL, "/") + "/api/v1/admin/history/rents"
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "history service unavailable"})
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response"})
			return
		}
		if resp.StatusCode != http.StatusOK {
			c.JSON(resp.StatusCode, gin.H{"error": "history service error"})
			return
		}
		var list []struct {
			ID            string  `json:"id"`
			ClientID      string  `json:"client_id"`
			MerchantID    string  `json:"merchant_id"`
			Cost          float32 `json:"cost"`
			Duration      int     `json:"duration"`
			StartedAt     *string `json:"started_at,omitempty"`
			EndedAt       *string `json:"ended_at,omitempty"`
			TemplateID    string  `json:"template_id"`
			TemplateTitle string  `json:"template_title"`
		}
		if err := json.Unmarshal(body, &list); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid history response"})
			return
		}
		ctx := c.Request.Context()
		out := make([]adminRentResponse, 0, len(list))
		for _, r := range list {
			clientName := ""
			merchantName := ""
			if getUsernameByID != nil {
				if r.ClientID != "" {
					clientName, _ = getUsernameByID(ctx, r.ClientID)
				}
				if r.MerchantID != "" {
					merchantName, _ = getUsernameByID(ctx, r.MerchantID)
				}
			}
			out = append(out, adminRentResponse{
				ID:            r.ID,
				ClientID:      r.ClientID,
				MerchantID:    r.MerchantID,
				ClientName:    clientName,
				MerchantName:  merchantName,
				Cost:          r.Cost,
				Duration:      r.Duration,
				StartedAt:     r.StartedAt,
				EndedAt:       r.EndedAt,
				TemplateID:    r.TemplateID,
				TemplateTitle: r.TemplateTitle,
			})
		}
		c.JSON(http.StatusOK, out)
	}
}

// proxyAdminTemplatesCreate проксирует POST /api/admin/templates в sessionhandlerservice.
func proxyAdminTemplatesCreate() gin.HandlerFunc {
	baseURL := os.Getenv("SESSION_HANDLER_HTTP_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8096"
	}
	client := &http.Client{Timeout: 15 * time.Second}
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
			return
		}
		permit := authMiddleware.GetPermitFromContext(c)
		if permit == nil || !permit.HasPermission(permission.SystemSettingsWrite) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: SystemSettingsWrite required"})
			return
		}
		body, _ := io.ReadAll(c.Request.Body)
		url := strings.TrimSuffix(baseURL, "/") + "/api/admin/templates"
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
			return
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "session handler unavailable"})
			return
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, "application/json", respBody)
	}
}

// proxyAdminTemplatesGet проксирует GET /api/admin/templates/:templateId в sessionhandlerservice (полный шаблон с ports, envs, volumes).
func proxyAdminTemplatesGet() gin.HandlerFunc {
	baseURL := os.Getenv("SESSION_HANDLER_HTTP_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8096"
	}
	client := &http.Client{Timeout: 15 * time.Second}
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
			return
		}
		permit := authMiddleware.GetPermitFromContext(c)
		if permit == nil || !permit.HasPermission(permission.SystemSettingsWrite) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: SystemSettingsWrite required"})
			return
		}
		templateId := c.Param("templateId")
		if templateId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "templateId required"})
			return
		}
		url := strings.TrimSuffix(baseURL, "/") + "/api/admin/templates/" + templateId
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "session handler unavailable"})
			return
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, "application/json", respBody)
	}
}

// proxyAdminTemplatesUpdate проксирует PATCH /api/admin/templates/:templateId в sessionhandlerservice.
func proxyAdminTemplatesUpdate() gin.HandlerFunc {
	baseURL := os.Getenv("SESSION_HANDLER_HTTP_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8096"
	}
	client := &http.Client{Timeout: 15 * time.Second}
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPatch {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
			return
		}
		permit := authMiddleware.GetPermitFromContext(c)
		if permit == nil || !permit.HasPermission(permission.SystemSettingsWrite) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: SystemSettingsWrite required"})
			return
		}
		templateId := c.Param("templateId")
		if templateId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "templateId required"})
			return
		}
		body, _ := io.ReadAll(c.Request.Body)
		url := strings.TrimSuffix(baseURL, "/") + "/api/admin/templates/" + templateId
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPatch, url, bytes.NewReader(body))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
			return
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "session handler unavailable"})
			return
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, "application/json", respBody)
	}
}
