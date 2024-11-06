package config

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/glebarez/sqlite"
)

func Database() *sql.DB {
	loadEnv()

	dbPath := os.Getenv("DBPATH")

	var err error

	db, err := sql.Open("sqlite", dbPath)

	if err != nil {
		log.Fatal(err)
	}

	return db
}
