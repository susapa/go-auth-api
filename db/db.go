package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init(dsn string) {
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	if err = DB.Ping(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	log.Println("database connected")
}

func Get() *sql.DB {
	return DB
}
