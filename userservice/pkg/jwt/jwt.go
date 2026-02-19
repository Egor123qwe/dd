package jwt

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrInvalidToken = fmt.Errorf("invalid token")
)

type JWT interface {
	Generate(id int64, perms uint64, exp time.Duration) (string, error)
	Validate(token string) (TokenClaims, error)
}

type TokenClaims struct {
	UserID       int
	EncodedPerms uint64
	TokenExp     time.Time
}

type JWTclaims struct {
	jwt.RegisteredClaims
	EncodedPerms uint64 `json:"roles"`
}

type service struct {
	secret string
}

func New(secret string) JWT {
	return &service{
		secret: secret,
	}
}

func (s *service) Generate(id int64, roles uint64, exp time.Duration) (string, error) {
	now := time.Now()

	claims := JWTclaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(id, 10),
			ExpiresAt: jwt.NewNumericDate(now.Add(exp)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},

		EncodedPerms: roles,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (s *service) Validate(tokenString string) (TokenClaims, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(s.secret), nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTclaims{}, keyFunc)
	if err != nil {
		return TokenClaims{}, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*JWTclaims)
	if !ok || !token.Valid {
		return TokenClaims{}, ErrInvalidToken
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return TokenClaims{}, fmt.Errorf("%w: invalid user ID", ErrInvalidToken)
	}

	result := TokenClaims{
		UserID:       int(userID),
		EncodedPerms: claims.EncodedPerms,
		TokenExp:     claims.ExpiresAt.Time,
	}

	return result, nil
}
