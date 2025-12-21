package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"database/sql"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	pb "github.com/yuramishin/expense-tracker/proto/pb_auth"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

type server struct {
	pb.UnimplementedAuthServiceServer
	db    *sql.DB
	redis *redis.Client
}

func (s *server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	hashedPass, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	var userID int64
	err := s.db.QueryRow("INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
		req.Email, string(hashedPass)).Scan(&userID)

	if err != nil {
		return nil, fmt.Errorf("could not register user: %v", err)
	}
	return &pb.RegisterResponse{UserId: userID}, nil
}

func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	var id int64
	var hash string
	err := s.db.QueryRow("SELECT id, password_hash FROM users WHERE email = $1", req.Email).Scan(&id, &hash)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("wrong password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": id,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString(jwtKey)

	return &pb.LoginResponse{Token: tokenString}, nil
}

func (s *server) Validate(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	exists, _ := s.redis.Exists(ctx, "blacklist:"+req.Token).Result()
	if exists > 0 {
		return &pb.ValidateResponse{Valid: false}, nil
	}

	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return &pb.ValidateResponse{Valid: false}, nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return &pb.ValidateResponse{Valid: false}, nil
	}

	return &pb.ValidateResponse{
		Valid:  true,
		UserId: int64(claims["user_id"].(float64)),
	}, nil
}

func (s *server) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	err := s.redis.Set(ctx, "blacklist:"+req.Token, "revoked", 24*time.Hour).Err()
	if err != nil {
		return &pb.LogoutResponse{Success: false}, err
	}
	return &pb.LogoutResponse{Success: true}, nil
}

func main() {
	db, err := sql.Open("postgres", "postgres://user:pass@postgres:5432/ledger?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	db.Exec(`CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, email TEXT UNIQUE, password_hash TEXT)`)

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

	lis, _ := net.Listen("tcp", ":50051")
	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, &server{db: db, redis: rdb})

	log.Println("Auth Service running on :50051 with Redis Blacklist")
	grpcServer.Serve(lis)
}
