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

type PipelineConfig struct {
	JobCh        chan string
	ResultCh     chan Result
	ErrCh        chan error
	ClientHandle *http.Client
	SrcFileName  string
	DstFileName  string
	HostName     string
}

// produce генерирует задания для обработки,
// комбинируя host со значениями из filename,
// и помещает их в канал outCh
func produce(cfg *PipelineConfig) {
	filename := cfg.SrcFileName
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
		cfg.JobCh <- "https://" + cfg.HostName + "/" + s
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "reading %s: %v\n", filename, err)
	}
}

// worker получает значения из канала inCh, пока он остается открытым,
// выполняет обработку и помещает результаты в outCh
func worker(cfg *PipelineConfig) {
	for job := range cfg.JobCh {
		resp, err := cfg.ClientHandle.Get(job)
		if err != nil {
			continue
		}
		resp.Body.Close()
		io.Copy(io.Discard, resp.Body)

		result := Result{
			Name: job,
			Code: resp.StatusCode,
		}
		cfg.ResultCh <- result
	}
}

func collect(cfg *PipelineConfig) {
	filename := cfg.DstFileName
	dstFile, err := os.Create(filename)
	if err != nil {
		cfg.ErrCh <- fmt.Errorf("creating %s: %w\n", filename, err)
	}
	defer dstFile.Close()

	writer := bufio.NewWriter(dstFile)

	for r := range cfg.ResultCh {
		s := fmt.Sprintf("%s - %d %s\n", r.Name, r.Code, http.StatusText(r.Code))
		_, err = writer.WriteString(s)
		if err != nil {
			cfg.ErrCh <- fmt.Errorf("writing to %s: %w\n", filename, err)
		}
	}

	if err := writer.Flush(); err != nil {
		cfg.ErrCh <- fmt.Errorf("writing to %s: %w\n", filename, err)
	}
}

func main() {
	const (
		srcFileName = "subdomains.txt"
		dstFileName = "results.txt"
		maxWorkers  = 20
	)

	// Целевой хост - аргумент
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "Target address not specified\n")
		os.Exit(1)
	}
	targetHost := os.Args[1]

	// Настроенный экземпляр HHTP-клиента
	client := &http.Client{
		Timeout: 1 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Игнорируем редиректы
			return http.ErrUseLastResponse
		},
	}

	// Канал с заданиями
	jobCh := make(chan string, maxWorkers)
	// Канал с результатами
	resultCh := make(chan Result, maxWorkers)
	// Канал с ошибками
	errCh := make(chan error, 3)

	config := &PipelineConfig{
		JobCh:        jobCh,
		ResultCh:     resultCh,
		ErrCh:        errCh,
		ClientHandle: client,
		HostName:     targetHost,
		SrcFileName:  srcFileName,
		DstFileName:  dstFileName,
	}

	// Группа конвейера обработки
	var pipelineWG sync.WaitGroup
	// Группа пула воркеров
	var workerWG sync.WaitGroup

	// Запускаем конвейер обработки
	pipelineWG.Go(func() {
		// collect собирает результаты из канала resultCh, пока он открыт
		collect(config)
	})
	pipelineWG.Go(func() {
		// produce вычитывает исходный файл до конца или ошибки и завершается, после чего закрывается канал jobCh
		produce(config)
		close(jobCh)
	})
	for range maxWorkers {
		// горутины пула воркеров работают до тех пор, пока открыт канал jobCh, после чего завершаются
		workerWG.Go(func() {
			worker(config)
		})
	}

	// Закрываем канал результатов по завершению пула воркеров
	go func() {
		// отдельная горутина ждет завершения пула воркеров, после этого закрывает канал resultCh и завершается сама
		workerWG.Wait()
		close(resultCh)
		// раз resultCh закрыт, то завершается и collect
	}()

	// Закрываем канал ошибок по завершению конвейера
	go func() {
		// collect и produce завершились, отдельная горутина, ожидающая их, закрывает канал errCh и завершается сама
		pipelineWG.Wait()
		close(errCh)
	}()

	// Сбор ошибок
	hasError := false
	for err := range errCh {
		if err != nil {
			fmt.Fprintf(os.Stderr, "error %v\n", err)
			hasError = true
		}
		// основной поток программы, в это время застрявший в цикле чтения из errCh, выходит из него
	}
	if hasError {
		os.Exit(1)
	}
}
