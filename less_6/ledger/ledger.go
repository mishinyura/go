package ledger

type Transaction struct {
	Category string
	Amount   float64
}

type Budget struct {
	Category string
	Limit    float64
	Period   string
}

type Ledger struct {
	transactions []Transaction
	budgets      map[string]Budget
}

func NewLedger() *Ledger {
	return &Ledger{
		transactions: []Transaction{},
		budgets:      make(map[string]Budget),
	}
}

func (l *Ledger) ListTransactions() []Transaction {
	out := make([]Transaction, len(l.transactions))
	copy(out, l.transactions)
	return out
}

func (l *Ledger) ListBudgets() []Budget {
	out := make([]Budget, 0, len(l.budgets))
	for _, b := range l.budgets {
		out = append(out, b)
	}
	return out
}

func (l *Ledger) Reset() {
	l.transactions = nil
	l.budgets = make(map[string]Budget)
}

func Reset(l *Ledger) { l.Reset() }
