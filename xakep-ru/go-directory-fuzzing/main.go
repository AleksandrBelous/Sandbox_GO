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
// комбинируя host со значениями, считанными из filename,
// и помещает их в канал outChannel.
func produce(filename string, host string, outChannel chan<- string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "opening %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	scanningDirs := bufio.NewScanner(file)

	for scanningDirs.Scan() {
		scanningDir := strings.TrimSpace(scanningDirs.Text())

		if scanningDir == "" {
			continue
		}

		outChannel <- "https://" + host + "/" + scanningDir
	}

	if err := scanningDirs.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "reading %s: %v\n", filename, err)
	}
}

// worker получает значения из канала inChannel, пока он остается открытым,
// выполняет обработку и помещает результаты в outChannel.
func worker(client *http.Client, inChannel <-chan string, outChannel chan<- Result) {
	for job := range inChannel {
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

		outChannel <- result
	}
}

// collect получает значения из канала resultChannel, пока он остается открытым,
// и записывает их в файл filename.
func collect(filename string, resultChannel <-chan Result) {
	dstFile, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "creating %s: %v\n", filename, err)
		return
	}
	defer dstFile.Close()

	writer := bufio.NewWriter(dstFile)

	for r := range resultChannel {
		s := fmt.Sprintf("%s - %d %s\n", r.Name, r.Code, http.StatusText(r.Code))
		_, err = writer.WriteString(s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "writing to %s: %v\n", filename, err)
		}
	}

	if err = writer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "writing to %s: %v\n", filename, err)
	}
}

func main() {
	// TODO
}
