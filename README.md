# dbwrap

Package to wrap & simplify (hopefully) access to a SQL database. Works on top of "database/sql" and not an ORM.
Features:

* Encapsulate low-level database access in a set of application-logic functions
* Prepared statements compiled on DB open time
* Transactions and embedded transactions

## basic example

Open and close. You should create a struct with a predefined field of type dbwrap.CleanupFunc and of name Cleanup (this is required). The field will be initialized during the call to Open and will hold all the code necessary to gracefully shut down all the initialized data 

```go
import (
	"log"
	"github.com/trueblacker/dbwrap"
)

type DB struct {
	Cleanup dbwrap.CleanupFunc
}

func (this *DB) Open() (*DB, error) {
	err := dbwrap.NewDB(this).Params(&dbwrap.DBParams{
		CreateReqs: dbwrap.QueryMap{
			dbwrap.DriverSQLite3: `
				CREATE TABLE IF NOT EXISTS user (
					id INTEGER PRIMARY KEY,
					name TEXT,
					data TEXT
				);
			`,
		},
	}).Open(dbwrap.DriverSQLite3, "file::memory:?mode=memory&cache=shared")
	if err != nil {
		return nil, err
	}
	return this, nil
}

func main() {
	db, err := new(DB).Open()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Cleanup()

}
```

## adding database access functions

To access the database dbwrap uses lazy-initialized function-typed struct fields, just like with Cleanup in the example above.
Each function field must have a specific user-defined type wich is basically is a function type but with a member function Prepare defined on that function (weird, huh?).

The Prepare function should implement the lazy initialization for the function itself (see *this = func...) which includes compilation of all the necessary prepared statements.

```go
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

type DB struct {
	Cleanup dbwrap.CleanupFunc
	AddUser dbFuncAddUser
}

func main() {
	db, err := new(DB).Open()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Cleanup()
	err = db.AddUser(context.Background(), "test_user")
	if err != nil {
		log.Fatal(err)
	}
}
```
