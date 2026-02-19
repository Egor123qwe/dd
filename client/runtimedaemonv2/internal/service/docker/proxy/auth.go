package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	auth "github.com/abbot/go-http-auth"
	proxyAuth "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/auth"
	"golang.org/x/crypto/bcrypt"
)

func (s service) createAuth(creeds proxyAuth.Credentials) auth.SecretProvider {
	return func(user, realm string) string {
		if user == creeds.Login {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creeds.Password), bcrypt.DefaultCost)

			if err == nil {
				return string(hashedPassword)
			}
		}

		return ""
	}
}

func (s service) createHandle(containerPort string) auth.AuthenticatedHandlerFunc {
	return func(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
		targetURL, err := url.Parse(fmt.Sprintf("http://localhost:%s", containerPort))

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		proxy.ServeHTTP(w, &r.Request)
	}
}
