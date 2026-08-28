package main

import (
	"fmt"
	"net/http"
)

func fetchData(url string) error {
	resp, err := http.Get(url)

	if err != nil {
		return fmt.Errorf("ошибка получения данных: %v", err)
	}

	defer resp.Body.Close()

	// обработка данных из ответа
	fmt.Println(resp.Status)
	fmt.Println(resp.Header)
	fmt.Println(resp.Body)

	return nil
}

func main() {
	err := fetchData("https://example.com")

	if err != nil {
		fmt.Println("Ошибка:", err)
	}
}
