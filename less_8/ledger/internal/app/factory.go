package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/you/monorepo/ledger/internal/domain"
	"github.com/you/monorepo/ledger/internal/repository/pg"
	"github.com/you/monorepo/ledger/internal/service"
)

type Config struct {
	DSN string
	// здесь можно добавить настройки кэша и тд
}

// NewServiceFactory создае и возвращает service.Service, closeFn, error
func NewServiceFactory(ctx context.Context, cfg Config) (service.Service, func(), error) {
	dsn := cfg.DSN
	if dsn == "" {
		dsn = defaultDSNFromEnv()
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("open db: %w", err)
	}
	// Ping с контекстом
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("ping db: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// репозитории
	repo := pg.NewRepo(db)

	// reporter: опционально - пока nil
	var reporter domain.Reporter = nil // если есть реализация, передать её

	// сервис
	svc := service.NewService(repo, repo, reporter) // repo реализует оба интерфейса

	closeFn := func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing db: %v", err)
		}
	}

	log.Println("service factory: db connected and service ready")
	return svc, closeFn, nil
}

func defaultDSNFromEnv() string {
	host := getenv("DB_HOST", "localhost")
	port := getenv("DB_PORT", "5432")
	user := getenv("DB_USER", "postgres")
	pass := getenv("DB_PASS", "postgres")
	name := getenv("DB_NAME", "cashapp")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, name)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
