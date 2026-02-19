package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func Generate(exp time.Duration, secret string) (string, error) {
	jwtClaims := jwt.MapClaims{
		"exp": time.Now().Add(exp).Unix(),
	}

	tokenJwt := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	token, err := tokenJwt.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return token, nil
}
