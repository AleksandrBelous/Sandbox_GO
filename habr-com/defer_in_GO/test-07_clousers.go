package main

import (
	"fmt"
)

func calculateSum(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Восстановление после паники:", r)
			err = fmt.Errorf("panic occurred")
		}
	}()

	result = a + b

	if result > 100 {
		panic("Сумма слишком велика")
	}

	return result, nil
}

func main() {
	sum, err := calculateSum(50, 60)

	if err != nil {
		fmt.Println("Ошибка:", err)
	} else {
		fmt.Println("Сумма:", sum)
	}
}
