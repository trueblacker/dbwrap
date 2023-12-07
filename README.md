# dbwrap

Package to wrap & simplify (hopefully) access to an SQL database. Works on top of "database/sql" and IS NOT an ORM.
Features:

* Encapsulate low-level database access in a set of application-logic functions
* Prepared statements compiled on DB open time
* Transactions and embedded transactions

## basic example

Open and close. Upon return, Open function returns the cleanup function, which holds all the code necessary to gracefully shut down the initialized data (close the database, etc).

Also notice dummy import of necessary database driver. The packagae itself does not implicitly imports any of them.

```go
import (
	"log"
	_ "github.com/mattn/go-sqlite3"
	"github.com/trueblacker/dbwrap"
)

type DB struct {
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
	}).Open(dbwrap.DriverSQLite3, "file::memory:?mode=memory&cache=shared")
	if err != nil {
		return nil, cleanup, err
	}
	return this, cleanup, nil
}

func main() {
	db, cleanup, err := new(DB).Open()
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()
}
```

## adding database access functions

To access the database dbwrap uses lazy-initialized function-typed struct fields, just like with Cleanup in the example above.
Each function field must have a specific user-defined type wich is basically is a function type but with a member function Prepare defined on that function (weird, huh?).

The Prepare function should implement the lazy initialization for the function itself (see *this = func...) which includes compilation of all the necessary prepared statements.

```go
type User struct {
	Name string
}

type dbFuncAddUser func(ctx context.Context, name string) (int, error)

func (this *dbFuncAddUser) Prepare(prms *dbwrap.Params) error {
	stmt, err := prms.Prepare(dbwrap.QueryMap{
		dbwrap.DriverSQLite3: `INSERT INTO user (name) VALUES (?) RETURNING id`,
	}[prms.Driver])
	if err != nil {
		return err
	}
	*this = func(ctx context.Context, name string) (int, error) {
		var id int
		err := stmt.QueryRowContext(ctx, name).Scan(&id)
		return id, err
	}
	return nil
}

type DB struct {
	AddUser dbFuncAddUser
}

func main() {
	db, cleanup, err := new(DB).Open()
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()
	ctx := context.Background()
	id, err := db.AddUser(ctx, "test_user")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("id: %v", id)
}
```
