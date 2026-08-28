package main

import (
	"fmt"
	"time"
)

func logFunctionTime() func() {
	// Особенность defer: откладывается не вычисление всего выражения после defer,
	// а сам вызов функции. Значения, необходимые для этого вызова,
	// вычисляются сразу в момент объявления defer.
	start := time.Now()

	// Возвращаем функцию, которая помнит start после выполнения тела defer,
	// - получили функцию-замыкание.
	return func() {
		fmt.Printf("Время выполнения: %v\n", time.Since(start))
	}
}

func funcA() {
	defer logFunctionTime()()

	time.Sleep(1 * time.Second)
	fmt.Println("Выполнение функции A")
}

func funcB() {
	defer logFunctionTime()()

	time.Sleep(500 * time.Millisecond)
	fmt.Println("Выполнение функции B")
}

func main() {
	funcA()
	funcB()
}
