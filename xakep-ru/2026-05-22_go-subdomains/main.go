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

type SubdomainResult struct {
	SubdomainName    string
	ReturnStatusCode int
}

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

	// Коллекция экземпляров типа SubdomainResult, которая будет
	// общим хранилищем результатов для всех горутин, - срез results
	results := []SubdomainResult{}

	// Создаём WaitGroup
	var wg sync.WaitGroup
	var mu sync.Mutex

	for scanner.Scan() {
		sub := strings.TrimSpace(scanner.Text())
		if sub == "" {
			continue
		}

		wg.Go(func() {
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

			fmt.Printf("*")
			result := SubdomainResult{
				SubdomainName:    targetHost,
				ReturnStatusCode: respStatusCode,
			}

			mu.Lock()
			defer mu.Unlock()
			results = append(results, result)
		})

		wg.Wait()
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading %s: %w", srcFileName, err)
	}

	for _, r := range results {
		fmt.Printf("%s - %d %s\n", r.SubdomainName, r.ReturnStatusCode, http.StatusText(r.ReturnStatusCode))
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
