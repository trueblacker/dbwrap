package dbwrap

import (
	"database/sql"
	"fmt"
)

type DriverName = string
type QueryStr = string

const (
	DriverSQLite3  = "sqlite3"
	DriverPostgres = "postgres"
)

type QueryMap map[DriverName]QueryStr

type CleanupFunc func()

type PrepareFunc func(string) (*sql.Stmt, error)
type TXFunc func(func(x *sql.Tx) error) error

type Params struct {
	Driver  string
	Prepare PrepareFunc
	tx      TXFunc
}

func (this *Params) PrepareByDriver(m QueryMap) (*sql.Stmt, error) {
	req, ok := m[this.Driver]
	if !ok || req == "" {
		return nil, fmt.Errorf("can't find statement for %v driver", this.Driver)
	}
	return this.Prepare(req)
}

func (this *Params) Tx(tx *sql.Tx) TXFunc {
	if tx != nil {
		return func(f func(tx *sql.Tx) error) error {
			return f(tx)
		}
	}
	return this.tx
}
