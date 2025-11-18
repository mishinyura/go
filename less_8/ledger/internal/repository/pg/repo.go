package pg

import (
	"database/sql"
	"time"

	"github.com/you/monorepo/ledger/internal/domain"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

// ------- BudgetRepository impl -------

func (r *Repo) UpsertBudget(b domain.Budget) error {
	const q = `
	INSERT INTO budgets(category, limit_amount)
	VALUES($1,$2)
	ON CONFLICT(category) DO UPDATE
	SET limit_amount = EXCLUDED.limit_amount`
	_, err := r.db.Exec(q, b.Category, b.Limit)
	return err
}

func (r *Repo) GetBudgetByCategory(category string) (domain.Budget, bool, error) {
	var b domain.Budget
	err := r.db.QueryRow(`SELECT id, category, limit_amount FROM budgets WHERE category=$1`, category).
		Scan(&b.ID, &b.Category, &b.Limit)
	if err == sql.ErrNoRows {
		return domain.Budget{}, false, nil
	}
	if err != nil {
		return domain.Budget{}, false, err
	}
	return b, true, nil
}

func (r *Repo) ListBudgets() ([]domain.Budget, error) {
	rows, err := r.db.Query(`SELECT id, category, limit_amount FROM budgets ORDER BY category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Budget
	for rows.Next() {
		var b domain.Budget
		if err := rows.Scan(&b.ID, &b.Category, &b.Limit); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, nil
}

// ------- ExpenseRepository impl -------

func (r *Repo) AddExpense(tx domain.Transaction) (int, error) {
	const q = `INSERT INTO expenses(amount, category, description, date) VALUES($1,$2,$3,$4) RETURNING id`
	var id int
	date := tx.Date
	if date.IsZero() {
		date = time.Now()
	}
	err := r.db.QueryRow(q, tx.Amount, tx.Category, tx.Description, date).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repo) ListExpenses() ([]domain.Transaction, error) {
	rows, err := r.db.Query(`SELECT id, amount, category, description, date FROM expenses ORDER BY date DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []domain.Transaction
	for rows.Next() {
		var t domain.Transaction
		if err := rows.Scan(&t.ID, &t.Amount, &t.Category, &t.Description, &t.Date); err != nil {
			return nil, err
		}
		res = append(res, t)
	}
	return res, nil
}

func (r *Repo) SumExpensesByCategory(category string) (float64, error) {
	var sum sql.NullFloat64
	err := r.db.QueryRow(`SELECT COALESCE(SUM(amount),0) FROM expenses WHERE category=$1`, category).Scan(&sum)
	if err != nil {
		return 0, err
	}
	if sum.Valid {
		return sum.Float64, nil
	}
	return 0, nil
}

// Compile-time checks: Repo implements interfaces
var _ domain.BudgetRepository = (*Repo)(nil)
var _ domain.ExpenseRepository = (*Repo)(nil)
