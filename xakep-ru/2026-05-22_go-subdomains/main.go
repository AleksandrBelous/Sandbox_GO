package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func run() error {
	const srcFileName = "subdomains.txt"

	// Целевой хост - аргумент запуска
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "Target address not specified\n")
		os.Exit(1)
	}
	targetHost := os.Args[1]

	// Открываем список поддоменов
	srcFile, err := os.Open(srcFileName)
	if err != nil {
		return fmt.Errorf("opening %s: %w", srcFileName, err)
	}
	defer srcFile.Close()

	// Настраиваем HTTP-клиент
	client := &http.Client{
		Timeout: 1 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Игнорируем редиректы
			return http.ErrUseLastResponse
		},
	}

	// Построчно считываем и проверяем домены
	scanner := bufio.NewScanner(srcFile)

	// Создаём WaitGroup
	var wg sync.WaitGroup

	for scanner.Scan() {
		sub := strings.TrimSpace(scanner.Text())
		if sub == "" {
			continue
		}

		// Увеличиваем счётчик горутин перед запуском
		// Перед запуском горутины сообщаем вызовом wg.Add(1),
		// что запускается одна горутина.
		wg.Add(1)
		go func() {
			// Уменьшаем счётчик горутин перед выходом
			// Внутри горутины нужно не забыть сообщить о ее завершении
			// вызовом wg.Done().
			defer wg.Done()

			targetHostURL := "https://" + sub + "." + targetHost

			resp, err := client.Get(targetHostURL)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			io.Copy(io.Discard, resp.Body)

			respStatusCode := resp.StatusCode
			if respStatusCode == http.StatusNotFound {
				// 404
				return
			}
			fmt.Printf("%s - %d %s\n", targetHostURL, respStatusCode, http.StatusText(respStatusCode))
		}()

		// Дожидаемся завершения всех горутин
		// Внутри WaitGroup находится атомарный счетчик.
		// Вызывая Add(), мы прибавляем к нему значение, переданное в аргументе.
		// Вызывая Done(), декрементируем счетчик, то есть уменьшаем его на 1.
		// Сам Wait() ждет, пока счетчик не станет равен 0.
		wg.Wait()
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading %s: %w", srcFileName, err)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
