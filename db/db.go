package db

import (
	"database/sql"
	"embed"
	"log"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

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
	runMigrations(DB)
}

func Get() *sql.DB {
	return DB
}

func runMigrations(db *sql.DB) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		log.Fatalf("failed to read migrations: %v", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		sql, err := migrationsFS.ReadFile("migrations/" + e.Name())
		if err != nil {
			log.Fatalf("failed to read migration %s: %v", e.Name(), err)
		}
		if _, err := db.Exec(string(sql)); err != nil {
			log.Fatalf("failed to run migration %s: %v", e.Name(), err)
		}
		log.Printf("migration applied: %s", e.Name())
	}
}
