package main

import (
	"errors"
	"fmt"
	"time"
)

// Transaction — структура для финансовой транзакции
type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        time.Time
}

// Хранилище транзакций (в памяти)
var transactions []Transaction

// AddTransaction — добавление новой транзакции
func AddTransaction(tx Transaction) error {
	if tx.Amount == 0 {
		return errors.New("сумма транзакции не может быть равна 0")
	}

	tx.ID = len(transactions) + 1
	if tx.Date.IsZero() {
		tx.Date = time.Now()
	}

	transactions = append(transactions, tx)
	return nil
}

// ListTransactions — возвращает список всех транзакций
func ListTransactions() []Transaction {
	return transactions
}

func main() {
	fmt.Println("Ledger service started")

	// Добавляем несколько транзакций для теста
	_ = AddTransaction(Transaction{Amount: 1200.50, Category: "Еда", Description: "Обед в кафе"})
	_ = AddTransaction(Transaction{Amount: 5000, Category: "Транспорт", Description: "Заправка"})
	_ = AddTransaction(Transaction{Amount: 250, Category: "Развлечения", Description: "Кино"})

	// Выводим все транзакции
	fmt.Println("\nСписок транзакций:")
	for _, tx := range ListTransactions() {
		fmt.Printf("ID: %d | %.2f руб | Категория: %s | %s | %s\n",
			tx.ID, tx.Amount, tx.Category, tx.Description, tx.Date.Format("2006-01-02 15:04:05"))
	}
}
