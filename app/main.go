package main

import (
	"fmt"
)

// func text(trans chan int) {
// 	fmt.Println("Начал")
// 	time.Sleep(5)
// 	fmt.Println("Закончил")
// 	trans <- 1
// }

func main() {
	// queue := make(chan int)
	// start := time.Now()
	// go text(queue)
	// go text(queue)
	// go text(queue)
	// go text(queue)
	// go text(queue)
	// result := queue
	// result = queue
	// result = queue
	// result = queue
	// result = queue
	// fmt.Println("Time:", time.Since(start), result)

	arr := []byte{1, 3, 2, 6, 3, 5, 7, 5, 3, 3, 5}

	for i, _ := range arr {
		fmt.Println(&arr[i])
		// fmt.Printf("%p\n", &arr)
	}
}
