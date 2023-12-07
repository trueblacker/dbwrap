# dbwrap

Package to wrap & simplify (hopefully) access to a SQL database. Works on top of "database/sql" and not an ORM.
Features:

* Encapsulate low-level database access in a set of application-logic functions
* Prepared statements compiled on DB open time
* Transactions and embedded transactions
