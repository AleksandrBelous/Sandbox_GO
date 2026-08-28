package main

import (
	"fmt"
	"time"
)

func logExecutionTime(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s заняла %s\n", name, elapsed)
}

func processData() {
	defer logExecutionTime(time.Now(), "processData")

	// типо обработка данных
	time.Sleep(2 * time.Second)
	fmt.Println("Данные обработаны")
}

func main() {
	processData()
}
