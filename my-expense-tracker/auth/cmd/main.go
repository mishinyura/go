package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
	pb "github.com/yuramishin/expense-tracker/proto/pb_auth"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

type server struct {
	pb.UnimplementedAuthServiceServer
	db *sql.DB
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateToken(userID int64, email string) (string, error) {
	if len(jwtKey) == 0 {
		jwtKey = []byte("default_secret_key") // Заглушка, если ENV не задан
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Токен живет 24 часа
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func (s *server) Register(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	hashedPass, err := hashPassword(req.Password)
	if err != nil {
		return &pb.AuthResponse{Error: "Failed to hash password"}, nil
	}

	var userID int64
	err = s.db.QueryRow("INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
		req.Email, hashedPass).Scan(&userID)

	if err != nil {
		return &pb.AuthResponse{Error: "User already exists or database error"}, nil
	}

	token, err := generateToken(userID, req.Email)
	if err != nil {
		return &pb.AuthResponse{Error: "Failed to generate token"}, nil
	}

	return &pb.AuthResponse{Token: token}, nil
}

func (s *server) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	var userID int64
	var hash string

	err := s.db.QueryRow("SELECT id, password_hash FROM users WHERE email = $1", req.Email).Scan(&userID, &hash)
	if err == sql.ErrNoRows {
		return &pb.AuthResponse{Error: "User not found"}, nil
	} else if err != nil {
		return &pb.AuthResponse{Error: "Database error"}, nil
	}

	if !checkPasswordHash(req.Password, hash) {
		return &pb.AuthResponse{Error: "Invalid credentials"}, nil
	}

	token, err := generateToken(userID, req.Email)
	if err != nil {
		return &pb.AuthResponse{Error: "Failed to generate token"}, nil
	}

	return &pb.AuthResponse{Token: token}, nil
}

func (s *server) Validate(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	if len(jwtKey) == 0 {
		jwtKey = []byte("default_secret_key")
	}

	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtKey, nil
	})

	if err != nil {
		return &pb.ValidateResponse{Valid: false}, nil
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := int64(claims["user_id"].(float64))
		return &pb.ValidateResponse{Valid: true, UserId: userID}, nil
	}

	return &pb.ValidateResponse{Valid: false}, nil
}

func main() {
	connStr := "postgres://user:pass@postgres:5432/ledger?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
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

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatal("Failed to create users table:", err)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, &server{db: db})

	log.Println("Auth Service running on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatal("Failed to serve:", err)
	}
}
