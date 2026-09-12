package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Result struct {
	Name string
	Code int
}

// produce генерирует задания для обработки,
// комбинируя host со значениями из filename,
// и помещает их в канал outCh
func produce(filename string, host string, outCh chan<- string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "opening %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		s := strings.TrimSpace(scanner.Text())
		if s == "" {
			continue
		}
		outCh <- "https://" + host + "/" + s
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "reading %s: %v\n", filename, err)
	}
}

// worker получает значения из канала inCh, пока он остается открытым,
// выполняет обработку и помещает результаты в outCh
func worker(client *http.Client, inCh <-chan string, outCh chan<- Result) {
	for job := range inCh {
		resp, err := client.Get(job)
		if err != nil {
			continue
		}
		resp.Body.Close()
		io.Copy(io.Discard, resp.Body)

		result := Result{
			Name: job,
			Code: resp.StatusCode,
		}
		outCh <- result
	}
}

func collect(filename string, resultCh <-chan Result) {
	dstFile, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "creating %s: %v\n", filename, err)
	}
	defer dstFile.Close()

	writer := bufio.NewWriter(dstFile)

	for r := range resultCh {
		s := fmt.Sprintf("%s - %d %s\n", r.Name, r.Code, http.StatusText(r.Code))
		_, err = writer.WriteString(s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "writing to %s: %v\n", filename, err)
		}
	}

	if err := writer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "writing to %s: %v\n", filename, err)
	}
}

func main() {
	// TODO
}
