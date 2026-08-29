package main

import (
	"fmt"
	"os"
)

func openFile(filename string) (f *os.File, err error) {
	f, err = os.Open(filename)

	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			f.Close()
			fmt.Println("Файл закрыт из-за ошибки")
		}
	}()

	// выполнение операций с файлом
	return f, nil
}

func main() {
	file, err := openFile("example.txt")

	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	defer file.Close()

	fmt.Println("Файл успешно открыт")
}
