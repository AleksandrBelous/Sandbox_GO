package main

import (
	"fmt"
)

func safeDivide(a, b int) int {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Восстановление после паники:", r)
		}
	}()

	return a / b
}

func main() {
	fmt.Println("Результат деления:", safeDivide(10, 0))
}
