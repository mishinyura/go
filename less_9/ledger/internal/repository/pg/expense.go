func (r *ExpenseRepo) GetCategories(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT DISTINCT category FROM expenses`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *ExpenseRepo) SumByCategory(ctx context.Context, category string, from, to time.Time) (float64, error) {
	var sum float64
	err := r.db.QueryRowContext(ctx, `
        SELECT COALESCE(SUM(amount), 0)
        FROM expenses
        WHERE category=$1 AND date BETWEEN $2 AND $3`, category, from, to).Scan(&sum)
	return sum, err
}
