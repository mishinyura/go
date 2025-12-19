package main

import (
	"context"
	"database/sql"
	"fmt"
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
	var limit float64
	err := s.db.QueryRow("SELECT limit_amount FROM budgets WHERE user_id = $1 AND category = $2", req.UserId, req.Category).Scan(&limit)

	if err == nil {
		var currentSpent float64
		s.db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE user_id = $1 AND category = $2", req.UserId, req.Category).Scan(&currentSpent)

		if currentSpent+req.Amount > limit {
			msg := fmt.Sprintf("Бюджет превышен! Лимит: %.0f, Уже потрачено: %.0f. Нельзя добавить %.0f", limit, currentSpent, req.Amount)
			log.Println(msg)
			return &pb.TransactionResponse{
				Success: false,
				Message: msg,
			}, nil
		}
	}

	_, err = s.db.Exec("INSERT INTO transactions (user_id, amount, category, description) VALUES ($1, $2, $3, $4)",
		req.UserId, req.Amount, req.Category, req.Description)

	if err != nil {
		log.Printf("Error inserting transaction: %v", err)
		return &pb.TransactionResponse{Success: false, Message: "Database error"}, nil
	}

	go s.cache.SetReport(context.Background(), req.UserId, nil)

	return &pb.TransactionResponse{Success: true, Message: "Saved"}, nil
}

func (s *server) GetReport(ctx context.Context, req *pb.ReportRequest) (*pb.ReportResponse, error) {
	if data, _ := s.cache.GetReport(ctx, req.UserId); data != nil {
		log.Println("Report Cache HIT")
		return &pb.ReportResponse{ByCategory: data}, nil
	}

	log.Println("🔍 Report Cache MISS")
	rows, err := s.db.Query("SELECT category, SUM(amount) FROM transactions WHERE user_id = $1 GROUP BY category", req.UserId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	report := make(map[string]float64)
	for rows.Next() {
		var cat string
		var sum float64
		rows.Scan(&cat, &sum)
		report[cat] = sum
	}

	go s.cache.SetReport(context.Background(), req.UserId, report)

	return &pb.ReportResponse{ByCategory: report}, nil
}

func (s *server) SetBudget(ctx context.Context, req *pb.BudgetRequest) (*pb.BudgetResponse, error) {
	_, err := s.db.Exec(`
		INSERT INTO budgets (user_id, category, limit_amount) 
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, category) DO UPDATE SET limit_amount = $3`,
		req.UserId, req.Category, req.LimitAmount)

	if err != nil {
		return &pb.BudgetResponse{Success: false, Message: "DB Error"}, nil
	}

	go s.cache.InvalidateBudgets(context.Background(), req.UserId)

	return &pb.BudgetResponse{Success: true, Message: "Budget Set"}, nil
}

func (s *server) GetBudgets(ctx context.Context, req *pb.GetBudgetsRequest) (*pb.BudgetList, error) {
	if cached, _ := s.cache.GetBudgets(ctx, req.UserId); cached != nil {
		log.Println("Budgets Cache HIT")
		return &pb.BudgetList{Budgets: cached}, nil
	}

	log.Println("🔍 Budgets Cache MISS")
	rows, err := s.db.Query("SELECT category, limit_amount FROM budgets WHERE user_id = $1", req.UserId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*pb.Budget
	for rows.Next() {
		b := &pb.Budget{}
		rows.Scan(&b.Category, &b.LimitAmount)
		list = append(list, b)
	}

	go s.cache.SetBudgets(context.Background(), req.UserId, list)

	return &pb.BudgetList{Budgets: list}, nil
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
		time.Sleep(2 * time.Second)
	}

	db.Exec(`CREATE TABLE IF NOT EXISTS transactions (
		id SERIAL PRIMARY KEY, user_id INT, amount FLOAT, category TEXT, description TEXT, created_at TIMESTAMP DEFAULT NOW())`)

	db.Exec(`CREATE TABLE IF NOT EXISTS budgets (
		id SERIAL PRIMARY KEY, user_id INT, category TEXT, limit_amount FLOAT, UNIQUE(user_id, category))`)

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient, _ := cache.NewRedisClient(redisAddr)

	lis, _ := net.Listen("tcp", ":50052")
	s := grpc.NewServer()
	pb.RegisterLedgerServiceServer(s, &server{db: db, cache: redisClient})

	log.Println("Ledger Service running on :50052")
	s.Serve(lis)
}
