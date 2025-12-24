package main

import (
	"database/sql"
	"log"
	"net"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	"github.com/yuramishin/expense-tracker/ledger/internal/config"
	"github.com/yuramishin/expense-tracker/ledger/internal/handler"
	"github.com/yuramishin/expense-tracker/ledger/internal/repository"
	"github.com/yuramishin/expense-tracker/ledger/internal/service"
	pb "github.com/yuramishin/expense-tracker/proto/pb_ledger"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	initDB(db)

	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	pgRepo := repository.NewPostgresRepo(db)
	redisRepo := repository.NewRedisRepo(rdb)
	svc := service.NewLedgerService(pgRepo, redisRepo)
	grpcHandler := handler.NewGrpcHandler(svc)

	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterLedgerServiceServer(s, grpcHandler)

	log.Printf("Ledger Service running on %s", cfg.GRPCPort)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func initDB(db *sql.DB) {
	db.Exec(`CREATE TABLE IF NOT EXISTS transactions (id SERIAL PRIMARY KEY, user_id INT, amount FLOAT, category TEXT, description TEXT, created_at TIMESTAMP DEFAULT NOW())`)
	db.Exec(`CREATE TABLE IF NOT EXISTS budgets (id SERIAL PRIMARY KEY, user_id INT, category TEXT, limit_amount FLOAT, UNIQUE(user_id, category))`)
}
