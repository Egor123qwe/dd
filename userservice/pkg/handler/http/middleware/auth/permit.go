package auth

import (
	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	"github.com/Interpuls/ifc2-service-farm/pkg/jwt"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
	"github.com/gin-gonic/gin"
	"strings"
)

const (
	PermitCtxKey = "permit"
)

func (m Middleware) getPermit(c *gin.Context) (permission.Permit, error) {
	permitFromCtx := GetPermitFromContext(c)

	if permitFromCtx != nil {
		return *permitFromCtx, nil
	}

	tokenClaimsData, err := m.validateAuthHeader(c.GetHeader("Authorization"))
	if err != nil {
		return permission.Permit{}, err
	}

	permit := permission.NewPermit(
		tokenClaimsData.UserID,
		permission.DecodePermissions(permission.EncodedPermission(tokenClaimsData.EncodedPerms))...,
	)

	return permit, nil
}

func GetPermitFromContext(c *gin.Context) *permission.Permit {
	tokenClaimsValue, ok := c.Get(PermitCtxKey)
	if ok {
		permit, ok := tokenClaimsValue.(permission.Permit)
		if ok {
			return &permit
		}
	}

	return nil
}

func (m Middleware) validateAuthHeader(header string) (jwt.TokenClaims, error) {
	if header == "" {
		return jwt.TokenClaims{}, errs.ErrUnauthorized
	}

	parts := strings.SplitN(header, " ", 2)

	if len(parts) != 2 || parts[0] != "Bearer" {
		return jwt.TokenClaims{}, errs.ErrUnauthorized
	}

	token := parts[1]

	claims, err := m.jwt.Validate(token)
	if err != nil {
		return jwt.TokenClaims{}, err
	}

	return claims, nil
}
