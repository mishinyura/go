package repository

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/yuramishin/expense-tracker/ledger/internal/domain"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) CreateTransaction(t *domain.Transaction) error {
	_, err := r.db.Exec("INSERT INTO transactions (user_id, amount, category, description) VALUES ($1, $2, $3, $4)",
		t.UserID, t.Amount, t.Category, t.Description)
	return err
}

func (r *PostgresRepo) GetBudget(userID int64, category string) (*domain.Budget, error) {
	var b domain.Budget
	err := r.db.QueryRow("SELECT category, limit_amount FROM budgets WHERE user_id = $1 AND category = $2", userID, category).Scan(&b.Category, &b.LimitAmount)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *PostgresRepo) GetTotalSpent(userID int64, category string) (float64, error) {
	var sum float64
	err := r.db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE user_id = $1 AND category = $2", userID, category).Scan(&sum)
	return sum, err
}

func (r *PostgresRepo) SetBudget(userID int64, category string, limit float64) error {
	_, err := r.db.Exec(`
		INSERT INTO budgets (user_id, category, limit_amount) VALUES ($1, $2, $3)
		ON CONFLICT (user_id, category) DO UPDATE SET limit_amount = $3`,
		userID, category, limit)
	return err
}

func (r *PostgresRepo) GetReportData(userID int64) (map[string]float64, error) {
	rows, err := r.db.Query("SELECT category, SUM(amount) FROM transactions WHERE user_id = $1 GROUP BY category", userID)
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
	return report, nil
}

func (r *PostgresRepo) GetBudgets(userID int64) ([]*domain.Budget, error) {
	rows, err := r.db.Query("SELECT category, limit_amount FROM budgets WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Budget
	for rows.Next() {
		b := &domain.Budget{}
		rows.Scan(&b.Category, &b.LimitAmount)
		list = append(list, b)
	}
	return list, nil
}
