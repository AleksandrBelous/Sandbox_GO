package main

import (
	"fmt"
	"sync"
)

func dangerFn() {
	defer func() {
		fmt.Println("topmost defer")
	}()

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered:", r)
		}
	}()

	defer func() {
		fmt.Println("another defer")
	}()

	fmt.Println("dangerous function")

	panic("oops!")

	defer func() {
		fmt.Println("unreachable defer")
	}()

	fmt.Println("unreachable code")
}

func main() {
	var wg sync.WaitGroup

	fmt.Println("start...")

	wg.Go(dangerFn)
	wg.Wait()

	fmt.Println("success!")
}

/*
Вывод:
start...
dangerous function
another defer
Recovered: oops! // Перехват паники
topmost defer
success!
*/
