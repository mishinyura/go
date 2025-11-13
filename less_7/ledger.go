package ledger

import (
	"database/sql"
	"errors"
	"time"

	"github.com/you/monorepo/ledger/internal/db"
)

type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        time.Time
}

type Budget struct {
	ID       int
	Category string
	Limit    float64
}

// валидация прежняя
func (t Transaction) Validate() error {
	if t.Category == "" {
		return errors.New("category cannot be empty")
	}
	if t.Amount == 0 {
		return errors.New("amount cannot be zero")
	}
	return nil
}
func (b Budget) Validate() error {
	if b.Category == "" {
		return errors.New("budget category cannot be empty")
	}
	if b.Limit <= 0 {
		return errors.New("budget limit must be greater than zero")
	}
	return nil
}

var ErrBudgetExceeded = errors.New("budget exceeded")

// --- работа с БД ---

func SetBudget(b Budget) error {
	if err := b.Validate(); err != nil {
		return err
	}
	const q = `
	INSERT INTO budgets(category, limit_amount)
	VALUES($1,$2)
	ON CONFLICT(category) DO UPDATE
	SET limit_amount = EXCLUDED.limit_amount`
	_, err := db.DB.Exec(q, b.Category, b.Limit)
	return err
}

func ListBudgets() ([]Budget, error) {
	rows, err := db.DB.Query(`SELECT id, category, limit_amount FROM budgets ORDER BY category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Budget
	for rows.Next() {
		var b Budget
		if err := rows.Scan(&b.ID, &b.Category, &b.Limit); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, nil
}

func AddTransaction(tx Transaction) error {
	if err := tx.Validate(); err != nil {
		return err
	}
	var limit sql.NullFloat64
	err := db.DB.QueryRow(`SELECT limit_amount FROM budgets WHERE category=$1`, tx.Category).Scan(&limit)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if limit.Valid {
		var spent float64
		_ = db.DB.QueryRow(`SELECT COALESCE(SUM(amount),0) FROM expenses WHERE category=$1`, tx.Category).Scan(&spent)
		if spent+tx.Amount > limit.Float64 {
			return ErrBudgetExceeded
		}
	}
	const insert = `INSERT INTO expenses(amount, category, description, date)
	                VALUES($1,$2,$3,$4)`
	_, err = db.DB.Exec(insert, tx.Amount, tx.Category, tx.Description, tx.Date)
	return err
}

func ListTransactions() ([]Transaction, error) {
	rows, err := db.DB.Query(`SELECT id, amount, category, description, date
		FROM expenses ORDER BY date DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []Transaction
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.Amount, &t.Category, &t.Description, &t.Date); err != nil {
			return nil, err
		}
		res = append(res, t)
	}
	return res, nil
}
