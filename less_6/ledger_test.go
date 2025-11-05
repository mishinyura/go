package ledger_test

import (
	"errors"
	"testing"

	"github.com/you/monorepo/ledger"
)

func TestTransactionValidate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		tx      ledger.Transaction
		wantErr bool
	}{
		{"ok", ledger.Transaction{Category: "еда", Amount: 100}, false},
		{"zero_amount", ledger.Transaction{Category: "еда", Amount: 0}, true},
		{"negative_amount", ledger.Transaction{Category: "еда", Amount: -1}, true},
		{"empty_category", ledger.Transaction{Category: "", Amount: 10}, true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.tx.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestBudgetValidate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		b       ledger.Budget
		wantErr bool
	}{
		{"ok", ledger.Budget{Category: "еда", Limit: 5000}, false},
		{"zero_limit", ledger.Budget{Category: "еда", Limit: 0}, true},
		{"negative_limit", ledger.Budget{Category: "еда", Limit: -10}, true},
		{"empty_category", ledger.Budget{Category: "", Limit: 100}, true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.b.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestAddTransaction_BudgetRules(t *testing.T) {
	l := ledger.NewLedger()
	t.Cleanup(func() { l.Reset() })

	if err := l.SetBudget(ledger.Budget{Category: "еда", Limit: 5000}); err != nil {
		t.Fatalf("set budget err: %v", err)
	}

	if err := l.AddTransaction(ledger.Transaction{Category: "еда", Amount: 1500}); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got := len(l.ListTransactions()); got != 1 {
		t.Fatalf("want 1 tx, got %d", got)
	}

	err := l.AddTransaction(ledger.Transaction{Category: "еда", Amount: 4000})
	if !errors.Is(err, ledger.ErrBudgetExceeded) {
		t.Fatalf("want ErrBudgetExceeded, got %v", err)
	}
	if got := len(l.ListTransactions()); got != 1 {
		t.Fatalf("want 1 tx after reject, got %d", got)
	}
}
