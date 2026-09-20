package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
)

type PipelineConfig struct {
	JobCh        chan string
	ResultCh     chan string
	DoneSignalCh <-chan struct{}
	ErrorCh      chan<- error
	SrcFileName  string
	TargetHash   *[md5.Size]byte
}

// produce генерирует задания для обработки, считывая значения из cfg.SrcFileName, и помещает их в канал cfg.JobCh;
// завершается по окончании чтения файла
func produce(cfg *PipelineConfig) {
	file, err := os.Open(cfg.SrcFileName)
	if err != nil {
		cfg.ErrorCh <- fmt.Errorf("opening %s^ %w", cfg.SrcFileName, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := strings.TrimSpace(scanner.Text())
		if s == "" {
			continue
		}

		select {
		case <-cfg.DoneSignalCh:
			return
		case cfg.JobCh <- s:
		}
	}

	if err := scanner.Err(); err != nil {
		cfg.ErrorCh <- fmt.Errorf("reading %s: %w", cfg.SrcFileName, err)
	}
}

// worker читает слова из cfg.JobCh, вычисляет MD5-хеш и сравнивает его с cfg.TargetHash;
// при совпадении помещает слово в cfg.ResultCh и завершается;
// при отсутствии совпадений завершается
func worker(cfg *PipelineConfig) {
	for {
		select {
		case job, ok := <-cfg.JobCh:
			if !ok {
				return
			}
			if md5.Sum([]byte(job)) == *cfg.TargetHash {
				cfg.ResultCh <- job
				return
			}
		case <-cfg.DoneSignalCh:
			return
		}
	}
}

// Функция collect ожидает чтения либо закрытия cfg.ResultCh, выводит значение в консоль и завершается
func collect(cfg *PipelineConfig) {
	pwd, ok := <-cfg.ResultCh
	if !ok {
		// Канал закрыт, данные не поступили
		fmt.Println("\rNo match found!")
	} else {
		// Принято сообщение из канала
		fmt.Println("\rMatched successfully: ", pwd)
	}
}

func main() {
	const srcFileName = "10k-most-common.txt"
	var maxWorkers = runtime.GOMAXPROCS(0)

	// Целевой хеш - аргумент запуска
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "Target MD5 hash not specified\n")
		os.Exit(1)
	}
	hashStr := strings.TrimSpace(os.Args[1])
	// Валидация
	if len(hashStr) != 32 {
		fmt.Fprintf(os.Stderr, "MD5 hash must be 32 chars long\n")
		os.Exit(1)
	}
	var hashBytes [md5.Size]byte
	_, err := hex.Decode(hashBytes[:], []byte(hashStr))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	// Канал с заданиями
	jobCh := make(chan string, maxWorkers)
	// Канал с результатами
	resultCh := make(chan string)
	// Канал с ошибками
	errCh := make(chan error, 3)
	// Канал для сигнализации о завершении
	doneSignalCh := make(chan struct{})

	config := &PipelineConfig{
		JobCh:        jobCh,
		ResultCh:     resultCh,
		DoneSignalCh: doneSignalCh,
		ErrorCh:      errCh,
		SrcFileName:  srcFileName,
		TargetHash:   &hashBytes,
	}

	// Группа конвейера обработки
	var piplineWg sync.WaitGroup
	// Группа пула воркеров
	var workerWg sync.WaitGroup

	// Запускаем конвейер обработки
	piplineWg.Go(func() {
		produce(config)
		close(jobCh)
	})
	piplineWg.Go(func() {
		collect(config)
		close(doneSignalCh)
	})
	// Запускаем пул горутин-обработчиков
	for range maxWorkers {
		workerWg.Go(func() {
			worker(config)
		})
	}

	// Закрываем канал результатов по завершении обработчиков
	go func() {
		workerWg.Wait()
		close(resultCh)
	}()

	// Закрываем канал ошибок по завершении конвейера
	go func() {
		piplineWg.Wait()
		close(errCh)
	}()

	// Сбор ошибок
	hadErr := false
	for err := range errCh {
		if err != nil {
			fmt.Fprintf(os.Stderr, "")
			hadErr = true
		}
	}
	if hadErr {
		os.Exit(1)
	}
}
