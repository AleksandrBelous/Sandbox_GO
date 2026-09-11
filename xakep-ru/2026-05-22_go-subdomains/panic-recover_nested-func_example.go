package main

import "fmt"

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
	dangerFn()
}

func main() {
	fmt.Println("start...")
	safeFn()
	fmt.Println("success!")
}

/*
Вывод:
start...
safe function
dangerous function
Recovered: oops! // Перехвачено
success!
*/
