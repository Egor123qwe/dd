package storage

import (
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/repo/balance"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/repo/cache"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/repo/rent"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/repo/session"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/repo/template"
)

func (s storage) Session() session.Session {
	return s.psql.Session()
}

func (s storage) Rent() rent.Rent {
	return s.psql.Rent()
}

func (s storage) Template() template.Template {
	return s.psql.Template()
}

func (s storage) Balance() balance.Balance {
	return s.psql.Balance()
}

func (s storage) Cache() cache.Repo {
	return s.redis.Cache()
}
