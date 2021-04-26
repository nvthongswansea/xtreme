package config

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)


func GetDB() (db *sql.DB, err error){
	db, err = sql.Open("sqlite3", "./db/sqlitextreme.db")
	return
}