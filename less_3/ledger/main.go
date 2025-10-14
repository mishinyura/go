package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	fmt.Println("=== Ledger demo ===")

	ledger := NewLedger()

	// Загружаем бюджеты из файла
	file, err := os.Open("budgets.json")
	if err != nil {
		fmt.Println("Ошибка при открытии файла бюджетов:", err)
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	if err := ledger.LoadBudgets(reader); err != nil {
		fmt.Println("Ошибка при загрузке бюджетов:", err)
		return
	}

	fmt.Println("Бюджеты успешно загружены")

	// Тестовые транзакции
	tests := []Transaction{
		{"еда", 1500},
		{"еда", 2000},
		{"еда", 1800},  // эта транзакция превысит лимит
		{"транспорт", 500},
	}

	for _, tx := range tests {
		err := ledger.AddTransaction(tx)
		if err != nil {
			fmt.Printf("Ошибка при добавлении транзакции [%s: %.2f]: %v\n", tx.Category, tx.Amount, err)
		} else {
			fmt.Printf("Транзакция [%s: %.2f] успешно добавлена\n", tx.Category, tx.Amount)
		}
	}

	ledger.PrintTransactions()
}