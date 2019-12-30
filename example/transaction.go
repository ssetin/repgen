package example

type Transaction interface {
	Commit() error
	Rollback() error
}
