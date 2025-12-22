package domain

import "time"

type Transaction struct {
	ID          int64
	UserID      int64
	Amount      float64
	Category    string
	Description string
	CreatedAt   time.Time
}

type Budget struct {
	ID          int64
	UserID      int64
	Category    string
	LimitAmount float64
}
