package dbwrap

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
)

type DBParams struct {
	CreateReqs QueryMap
	Funcs      []any
}

type dbparams struct {
	prms     DBParams
	funcs    []any
	preparer preparer
}

type dbOpener interface {
	Params(prms *DBParams) dbOpener
	Open(driver, connectstr string) error
}

func (this *dbparams) Params(prms *DBParams) dbOpener {
	this.prms = *prms
	return this
}

func NewDB[T any](ret *T) dbOpener {
	return &dbparams{
		funcs: Funcs(ret),
		preparer: &dbFuncsPreparer[T]{
			target: ret,
		},
	}
}

func (this *dbparams) Open(driver, connectstr string) error {
	db, err := ConnectDBWithPing(driver, connectstr)
	if err != nil {
		return err
	}
	if this.prms.CreateReqs != nil {
		req, ok := this.prms.CreateReqs[driver]
		if !ok || req == "" {
			return fmt.Errorf("can't find migrate statements for %v driver", driver)
		}
		_, err = db.Exec(req)
		if err != nil {
			return err
		}
	}
	err = this.prepareFuncs(db, driver, append(this.funcs, this.prms.Funcs...))
	if err != nil {
		return err
	}
	return nil
}

type stmts []*sql.Stmt

func (this stmts) getCleanup(db *sql.DB) func() {
	return func() {
		for _, s := range this {
			s.Close()
		}
		if db != nil {
			db.Close()
		}
	}
}

func (this dbparams) prepareFuncs(db *sql.DB, driver string, allFuncs []any) error {
	var allStmts stmts
	cleanup := allStmts.getCleanup(db)
	failCleanup := cleanup
	defer func() {
		if failCleanup != nil {
			failCleanup()
		}
	}()
	if len(allFuncs) > 0 {
		prms := &Params{
			Driver: driver,
			Prepare: func(req string) (*sql.Stmt, error) {
				stmt, err := db.Prepare(req)
				if err != nil {
					return nil, err
				}
				allStmts = append(allStmts, stmt)
				return stmt, nil
			},
			tx: func(f func(tx *sql.Tx) error) error {
				tx, err := db.Begin()
				if err != nil {
					return err
				}
				cleanup := tx.Rollback
				defer func() {
					if cleanup != nil {
						cleanup()
					}
				}()
				err = f(tx)
				if err != nil {
					return err
				}
				cleanup = nil
				return tx.Commit()
			},
		}

		check := func(v reflect.Value) bool {
			return v.Kind() == reflect.Func &&
				v.CanInterface() && v.IsNil() && v.CanSet()
		}
		for _, funcs := range allFuncs {
			ptr := reflect.ValueOf(funcs)
			if ptr.Kind() != reflect.Ptr {
				return errors.New("invalid funcs kind: ptr wanted")
			}
			f := ptr.Elem()
			if f.Kind() != reflect.Struct {
				return errors.New("invalid funcs type: ptr to struct wanted")
			}
			numField := f.NumField()
			for i := 0; i < numField; i++ {
				field := f.Field(i)
				if !check(field) {
					continue
				}
				err := this.preparer.prepare(field.Addr().Interface(), prms)
				if err != nil {
					return fmt.Errorf("error preparing sql statement for field %v: %v", i, err)
				}
			}
			field := f.FieldByName("Cleanup")
			if check(field) {
				if _, ok := field.Interface().(CleanupFunc); ok {
					field.Set(reflect.ValueOf(cleanup))
				}
			}
		}
	}
	failCleanup = nil
	return nil
}

func ConnectDBWithPing(driver, connectstr string) (*sql.DB, error) {
	db, err := sql.Open(driver, connectstr)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}
