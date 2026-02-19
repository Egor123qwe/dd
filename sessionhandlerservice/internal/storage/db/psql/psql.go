package psql

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	balancerepo "gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/db/psql/repo/balance"
	rentrepo "gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/db/psql/repo/rent"
	sessionrepo "gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/db/psql/repo/session"
	templaterepo "gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/db/psql/repo/template"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/repo/balance"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/repo/rent"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/repo/session"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/repo/template"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/pkg/sqlt"
)

type Store interface {
	Session() session.Session
	Rent() rent.Rent
	Template() template.Template
	Balance() balance.Balance

	Close() error
}

type store struct {
	db       *sqlx.DB
	rent     rent.Rent
	session  session.Session
	template template.Template
	balance  balance.Balance
}

func configure(stateMachineDB, templatesDB *sqlx.DB) Store {
	return store{
		db:       stateMachineDB,
		rent:     rentrepo.New(sqlt.NewDB(stateMachineDB)),
		session:  sessionrepo.New(stateMachineDB),
		template: templaterepo.New(templatesDB),
		balance:  balancerepo.New(stateMachineDB),
	}
}

func New(config Config) (Store, error) {
	stateMachineDB, err := sqlx.Connect(config.StateMachine.Driver, config.StateMachine.URL)
	if err != nil {
		return nil, err
	}

	templatesDB, err := sqlx.Connect(config.Templates.Driver, config.Templates.URL)
	if err != nil {
		return nil, err
	}

	return configure(stateMachineDB, templatesDB), nil
}

func (s store) Close() error {
	return s.db.Close()
}

func (s store) Rent() rent.Rent {
	return s.rent
}

func (s store) Session() session.Session {
	return s.session
}

func (s store) Template() template.Template {
	return s.template
}

func (s store) Balance() balance.Balance {
	return s.balance
}
