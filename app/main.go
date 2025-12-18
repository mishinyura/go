package main

import "fmt"

//Поверяем пользователя
func checkUser() (string, bool) {
	return "admin", true
}

func main() {
	role, ok := checkUser()

	{
		fmt.Println(role, ok)
	}

	fmt.Println(role, ok)
}
