package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/yuramishin/expense-tracker/ledger/internal/cache"
	pb "github.com/yuramishin/expense-tracker/proto/pb_ledger"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedLedgerServiceServer
	db    *sql.DB
	cache *cache.RedisClient
}

func (s *server) CreateTransaction(ctx context.Context, req *pb.TransactionRequest) (*pb.TransactionResponse, error) {
	_, err := s.db.Exec("INSERT INTO transactions (user_id, amount, category, description) VALUES ($1, $2, $3, $4)",
		req.UserId, req.Amount, req.Category, req.Description)

	if err != nil {
		log.Printf("Error inserting transaction: %v", err)
		return &pb.TransactionResponse{Success: false, Message: "Database error"}, nil
	}

	return &pb.TransactionResponse{Success: true, Message: "Saved"}, nil
}

func (s *server) GetReport(ctx context.Context, req *pb.ReportRequest) (*pb.ReportResponse, error) {
	if data, _ := s.cache.GetReport(ctx, req.UserId); data != nil {
		log.Println("🚀 Cache HIT (from Redis)")
		return &pb.ReportResponse{ByCategory: data}, nil
	}

	log.Println("🔍 Cache MISS (calculating from DB...)")
	rows, err := s.db.Query("SELECT category, SUM(amount) FROM transactions WHERE user_id = $1 GROUP BY category", req.UserId)
	if err != nil {
		log.Printf("Error querying report: %v", err)
		return nil, err
	}
	defer rows.Close()

	report := make(map[string]float64)
	for rows.Next() {
		var cat string
		var sum float64
		if err := rows.Scan(&cat, &sum); err != nil {
			continue
		}
		report[cat] = sum
	}

	go s.cache.SetReport(context.Background(), req.UserId, report)

	return &pb.ReportResponse{ByCategory: report}, nil
}

func main() {
	db, err := sql.Open("postgres", "postgres://user:pass@postgres:5432/ledger?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		log.Println("Waiting for database...")
		time.Sleep(2 * time.Second)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS transactions (
		id SERIAL PRIMARY KEY,
		user_id INT,
		amount FLOAT,
		category TEXT,
		description TEXT,
		created_at TIMESTAMP DEFAULT NOW()
	)`)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
	log.Println("Database table ensured")

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient, err := cache.NewRedisClient(redisAddr)
	if err != nil {
		log.Printf("Redis connection failed: %v", err)
	}

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatal(err)
	}
	s := grpc.NewServer()
	pb.RegisterLedgerServiceServer(s, &server{db: db, cache: redisClient})

	log.Println("Ledger Service running on :50052")
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
