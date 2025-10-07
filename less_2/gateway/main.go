package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
			return
		}
		fmt.Fprintln(w, "pong")
	})

	fmt.Println("Gateway server started on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
	}
}
