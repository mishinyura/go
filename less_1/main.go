package main

import (
	"fmt"
	"os"
	"runtime"
)

func main() {
	// Получаем имя пользователя из переменной окружения USER (Linux/macOS) или USERNAME (Windows)
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	fmt.Println("Имя пользователя:", user)

	// Читаем аргументы командной строки
	args := os.Args[1:] // первый элемент os.Args — это имя программы
	fmt.Println("Аргументы CLI:", args)

	// Текущая версия Go
	fmt.Println("Версия Go:", runtime.Version())
}
