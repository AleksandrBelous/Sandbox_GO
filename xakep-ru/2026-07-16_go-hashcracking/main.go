package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"os"
	"strings"
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
	// TODO
}
