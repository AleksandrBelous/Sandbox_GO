package main

import (
	"fmt"
	"time"
)

func handlePanic() {
	if r := recover(); r != nil {
		fmt.Println("Recovered:", r)
	}
}

func dangerFn() {
	fmt.Println("dangerous function")
	panic("oops!")
}

func safeFn() {
	defer handlePanic()
	fmt.Println("safe function")
	go dangerFn()
	time.Sleep(2 * time.Second)
}

func main() {
	fmt.Println("start...")
	go safeFn()
	time.Sleep(2 * time.Second)
	fmt.Println("success!")
}

/*
Вывод:
start...
safe function
dangerous function
panic: oops! // Аварийное завершение
*/
