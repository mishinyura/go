package repository

import (
	"context"
	"errors"
	"sync"
	"time"

	"gitlab.com/education/users-api/internal/model"
)

// Интерфейс репозитория пользователя
var (
	ErrUserNotFound = errors.New("user not found")
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	UpdatePartial(ctx context.Context, id string, patch model.UpdateUserRequest) (*model.User, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]model.User, error)
}

// Реализация на базе in-memory hashmap, потокобезопасная
type InMemoryUserRepository struct {
	mu    sync.RWMutex
	store map[string]model.User
}

func NewInMemoryUserRepository(seed map[string]model.User) *InMemoryUserRepository {
	if seed == nil {
		seed = make(map[string]model.User)
	}
	// Добавляем дефолтного пользователя, если его нет
	const defaultID = "11111111-1111-1111-1111-111111111111"
	if _, exists := seed[defaultID]; !exists {
		email := "alice@example.com"
		now := time.Now().UTC()
		seed[defaultID] = model.User{
			ID:        defaultID,
			Username:  "alice",
			Email:     &email,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	return &InMemoryUserRepository{store: seed}
}

func (r *InMemoryUserRepository) Create(ctx context.Context, user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	u := *user
	r.store[user.ID] = u

	return nil
}

func (r *InMemoryUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.store[id]
	if !ok {
		return nil, ErrUserNotFound
	}

	uCopy := u

	return &uCopy, nil
}

func (r *InMemoryUserRepository) UpdatePartial(ctx context.Context, id string, patch model.UpdateUserRequest) (*model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.store[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	if patch.Username != nil {
		u.Username = *patch.Username
	}
	if patch.Email != nil {
		u.Email = patch.Email
	}
	u.UpdatedAt = time.Now().UTC()
	r.store[id] = u
	uCopy := u

	return &uCopy, nil
}

func (r *InMemoryUserRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[id]; !ok {
		return ErrUserNotFound
	}
	delete(r.store, id)

	return nil
}

func (r *InMemoryUserRepository) List(ctx context.Context) ([]model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]model.User, 0, len(r.store))
	for _, u := range r.store {
		res = append(res, u)
	}

	return res, nil
}
