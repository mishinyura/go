package repository

import (
	"database/sql"

	"github.com/yuramishin/expense-tracker/auth/internal/domain"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) CreateUser(email, hash string) (int64, error) {
	var id int64
	err := r.db.QueryRow("INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id", email, hash).Scan(&id)
	return id, err
}

func (r *PostgresRepo) GetUserByEmail(email string) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow("SELECT id, password_hash FROM users WHERE email = $1", email).Scan(&u.ID, &u.PasswordHash)
	if err != nil {
		return nil, err
	}
	return u, nil
}
