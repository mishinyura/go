package leedcode

import (
	"fmt"
	"strconv"
)

func IsPalindromeStr(x int) bool {
	if 11 > x && x > -11 {
		return false
	}
	numStr := strconv.Itoa(x)
	amount := len(numStr)

	lCur := 0
	rCur := amount - 1

	for i := 0; i < amount/2; i++ {
		fmt.Println(numStr, lCur, rCur)
		if numStr[lCur] != numStr[rCur] {
			return false
		}
		lCur++
		rCur--
	}
	return true
}

func IsPalindromeNum(x int) bool {
	if x < 0 {
		return false
	}

	if x < 10 {
		return true
	}

	div := 1

	{
		temp := x
		for {
			if temp < 10 {
				break
			}
			div *= 10
			temp = temp / 10
		}
	}

	for i := 0; i < 10; i++ {
		sNum := x / div
		eNum := x % 10
		fmt.Println(sNum, eNum)
		x = x / 100
		x = x % 10 / 100
	}

	// fmt.Println(eNum, sNum, div)

	return true
}
