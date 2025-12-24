package main

import (
	"database/sql"
	"log"
	"net"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	"github.com/yuramishin/expense-tracker/auth/internal/config"
	"github.com/yuramishin/expense-tracker/auth/internal/handler"
	"github.com/yuramishin/expense-tracker/auth/internal/repository"
	"github.com/yuramishin/expense-tracker/auth/internal/service"
	pb "github.com/yuramishin/expense-tracker/proto/pb_auth"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 15; i++ {
		if err := db.Ping(); err == nil {
			log.Println("DB Connected")
			break
		}
		time.Sleep(2 * time.Second)
	}

	db.Exec(`CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, email TEXT UNIQUE, password_hash TEXT)`)

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
	authSvc := service.NewAuthService(pgRepo, redisRepo, cfg.JWTSecret)
	grpcHandler := handler.NewGrpcHandler(authSvc)

	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, grpcHandler)

	log.Printf("Auth Service running on %s", cfg.GRPCPort)
	s.Serve(lis)
}
