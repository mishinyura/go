package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

func Init() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		host := getenv("DB_HOST", "localhost")
		port := getenv("DB_PORT", "5432")
		user := getenv("DB_USER", "postgres")
		pass := getenv("DB_PASS", "postgres")
		name := getenv("DB_NAME", "cashapp")
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, name)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("cannot open DB: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("cannot connect DB: %v", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	DB = db
	log.Println("✅ connected to PostgreSQL")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
