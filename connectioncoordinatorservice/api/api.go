package api

import "gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/api/auth"

type Service interface {
	Auth() auth.Auth
}

type service struct {
	auth auth.Auth
}

func New() (Service, error) {
	srv := service{
		auth: auth.New(),
	}

	return srv, nil
}

func (s service) Auth() auth.Auth {

	return s.auth
}
