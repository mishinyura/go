package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yuramishin/expense-tracker/auth/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	pg        *repository.PostgresRepo
	redis     *repository.RedisRepo
	jwtSecret []byte
}

func NewAuthService(pg *repository.PostgresRepo, redis *repository.RedisRepo, secret string) *AuthService {
	return &AuthService{
		pg:        pg,
		redis:     redis,
		jwtSecret: []byte(secret),
	}
}

func (s *AuthService) Register(email, password string) (int64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	return s.pg.CreateUser(email, string(hash))
}

func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.pg.GetUserByEmail(email)
	if err != nil {
		return "", errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("wrong password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) Validate(ctx context.Context, tokenStr string) (int64, bool) {
	isBlacklisted, err := s.redis.IsBlacklisted(ctx, tokenStr)
	if err != nil {
		return 0, false
	}
	if isBlacklisted {
		return 0, false
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return 0, false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, false
	}

	return int64(claims["user_id"].(float64)), true
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.redis.AddToBlacklist(ctx, token, 24*time.Hour)
}
