package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	fmt.Println("=== Ledger demo (ООП, интерфейсы, валидация) ===")

	ledger := NewLedger()

	file, err := os.Open("budgets.json")
	if err != nil {
		fmt.Println("Ошибка при открытии файла:", err)
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	if err := ledger.LoadBudgets(reader); err != nil {
		fmt.Println("Ошибка при загрузке бюджетов:", err)
		return
	}
	fmt.Println("Бюджеты успешно загружены")

	fmt.Println("\n=== Проверка валидации ===")
	validTx := Transaction{Category: "еда", Amount: 1000}
	invalidTx := Transaction{Category: "", Amount: -500}
	validBudget := Budget{Category: "транспорт", Limit: 2000}
	invalidBudget := Budget{Category: "", Limit: -100}

	PrintValidation(validTx)
	PrintValidation(invalidTx)
	PrintValidation(validBudget)
	PrintValidation(invalidBudget)

	fmt.Println("\n=== Добавление транзакций ===")

	tests := []Transaction{
		{"еда", 1500},
		{"еда", 4000}, // превысит лимит
		{"", 1000},    // невалидная категория
	}

	for _, tx := range tests {
		err := ledger.AddTransaction(tx)
		if err != nil {
			fmt.Printf("Ошибка при добавлении [%s: %.2f]: %v\n", tx.Category, tx.Amount, err)
		} else {
			fmt.Printf("Добавлено: [%s: %.2f]\n", tx.Category, tx.Amount)
		}
	}

	ledger.PrintTransactions()
}