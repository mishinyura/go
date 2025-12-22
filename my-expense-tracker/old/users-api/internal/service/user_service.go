package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gitlab.com/education/users-api/internal/model"
	"gitlab.com/education/users-api/internal/repository"
)

// Сервис бизнес-логики пользователя
type UserService interface {
	CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.User, error)
	GetUser(ctx context.Context, id string) (*model.User, error)
	UpdateUser(ctx context.Context, id string, req model.UpdateUserRequest) (*model.User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context) ([]model.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
	now := time.Now().UTC()
	u := &model.User{
		ID:        uuid.NewString(),
		Username:  req.Username,
		Email:     req.Email,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *userService) GetUser(ctx context.Context, id string) (*model.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userService) UpdateUser(ctx context.Context, id string, req model.UpdateUserRequest) (*model.User, error) {
	return s.repo.UpdatePartial(ctx, id, req)
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *userService) ListUsers(ctx context.Context) ([]model.User, error) {
	return s.repo.List(ctx)
}
