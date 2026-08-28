package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	file, err := os.Open("example.txt")

	if err != nil {
		fmt.Println("Ошибка при открытии файла:", err)
		return
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Ошибка чтения файла:", err)
	}
}
