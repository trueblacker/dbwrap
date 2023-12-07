package dbwrap_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/trueblacker/dbwrap"
)

type User struct {
	Id   uint64
	Name string
	Data []byte
}

type dbFuncAddUser func(ctx context.Context, name string) error

func (this *dbFuncAddUser) Prepare(prms *dbwrap.Params) error {
	stmt, err := prms.Prepare(dbwrap.QueryMap{
		dbwrap.DriverSQLite3: `INSERT INTO user (name, data) VALUES (?, ?)`,
	}[prms.Driver])
	if err != nil {
		return err
	}
	*this = func(ctx context.Context, name string) error {
		return prms.Tx(nil)(func(tx *sql.Tx) error {
			_, err := tx.Stmt(stmt).ExecContext(ctx, name, dbwrap.Bin2Str([]byte("test")))
			return err
		})
	}
	return nil
}

type dbFuncGetUsers func(ctx context.Context) ([]*User, error)

func (this *dbFuncGetUsers) Prepare(prms *dbwrap.Params) error {
	stmt, err := prms.Prepare(`SELECT id, name, data FROM user`)
	if err != nil {
		return err
	}
	*this = func(ctx context.Context) ([]*User, error) {
		rows, err := stmt.QueryContext(ctx)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		ret := []*User{}
		for rows.Next() {
			var user User
			err = dbwrap.Query().Dests(
				&user.Id,
				&user.Name,
				&user.Data).Scan(rows)
			if err != nil {
				return nil, err
			}
			ret = append(ret, &user)
		}
		return ret, nil
	}
	return nil
}

type DB struct {
	GetUsers dbFuncGetUsers
	priv     struct {
		AddUser dbFuncAddUser
	}
}

func (this *DB) Open() (*DB, dbwrap.CleanupFunc, error) {
	cleanup, err := dbwrap.NewDB(this).Params(&dbwrap.DBParams{
		CreateReqs: dbwrap.QueryMap{
			dbwrap.DriverSQLite3: `
				CREATE TABLE IF NOT EXISTS user (
					id INTEGER PRIMARY KEY,
					name TEXT,
					data TEXT
				);
			`,
		},
		Funcs: dbwrap.Funcs(&this.priv),
	}).Open(dbwrap.DriverSQLite3, "file::memory:?mode=memory&cache=shared")
	if err != nil {
		return nil, cleanup, err
	}
	return this, cleanup, nil
}

func js(v any) string {
	ret, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(ret)
}

func Example() {
	db, cleanup, err := new(DB).Open()
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	ctx := context.Background()

	err = db.priv.AddUser(ctx, "test_user")
	if err != nil {
		log.Fatal(err)
	}

	users, err := db.GetUsers(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(js(users))

	// Output: [{"Id":1,"Name":"test_user","Data":"dGVzdA=="}]
}
