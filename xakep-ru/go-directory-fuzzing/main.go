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

	for resultRaw := range resultChannel {
		resultStr := fmt.Sprintf("%s - %d %s\n", resultRaw.Name, resultRaw.Code, http.StatusText(resultRaw.Code))
		_, err = writer.WriteString(resultStr)
		if err != nil {

			fmt.Fprintf(os.Stderr, "writing to %s: %v\n", filename, err)
		}
	}

	if err = writer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "writing to %s: %v\n", filename, err)
	}
}

func main() {
	const (
		srcFileName = "Common-DB-Backups.txt"
		dstFileName = "results.txt"
		maxWorkers  = 20
	)

	// Целевой хост - аргумент запуска
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "Target address not specified\n")
		os.Exit(1)
	}
	host := os.Args[1]

	// Настроенный экземпляр HTTP клиента
	client := &http.Client{
		Timeout: 1 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Игнорируем редиректы
			return http.ErrUseLastResponse
		},
	}

	// Канал с заданиями
	jobChannel := make(chan string, maxWorkers)

	// Канал с результатами
	resultChannel := make(chan Result, maxWorkers)

	// Группа конвейера обработки
	var pipelineWG sync.WaitGroup

	// Запускаем конвейер обработки
	pipelineWG.Go(func() {
		collect(dstFileName, resultChannel)
	})
	pipelineWG.Go(func() {
		produce(srcFileName, host, jobChannel)
		close(jobChannel)
	})

	// Группа пула воркеров
	var workerWG sync.WaitGroup

	// Запускаем пул воркеров
	for range maxWorkers {
		workerWG.Go(func() {
			worker(client, jobChannel, resultChannel)
		})
	}
	// Ожидаем завершения пула воркеров и закрываем канал результатов
	workerWG.Wait()
	close(resultChannel)

	// Ожидаем завершения конвейера
	pipelineWG.Wait()
}
