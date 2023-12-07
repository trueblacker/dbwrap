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
