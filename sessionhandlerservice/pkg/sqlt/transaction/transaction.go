package transaction

type Service interface {
	Rollback() error
	Commit() error
}
