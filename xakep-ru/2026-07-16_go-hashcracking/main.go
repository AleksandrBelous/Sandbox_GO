package main

import (
	"bufio"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type PipelineConfig struct {
	JobCh               chan []string
	ResultCh            chan string
	ErrorCh             chan<- error
	SrcFileName         string
	TargetHash          *[md5.Size]byte
	Counter             *atomic.Uint64
	CancelCheckInterval int
	BatchSize           int
}

// produce генерирует задания для обработки, считывая значения из cfg.SrcFileName, и помещает их в канал cfg.JobCh;
// завершается по окончании чтения файла
func produce(ctx context.Context, cfg *PipelineConfig) {
	file, err := os.Open(cfg.SrcFileName)
	if err != nil {
		cfg.ErrorCh <- fmt.Errorf("opening %s^ %w", cfg.SrcFileName, err)
		return
	}
	defer file.Close()

	// Пакет-задание, содержащий строки для обработки
	batch := make([]string, 0, cfg.BatchSize)
	// Счётчик считанных строк
	i := 0
	// Неблокирующая отправка пакета в канал cfg.JobCh:
	// возвращает true при успешном завершении записи
	// или false, если произошла отмена контекста
	sendBatch := func(batch []string) bool {
		select {
		case <-ctx.Done():
			return false
		case cfg.JobCh <- batch:
			return true
		}
	}

	// Считываем файл построчно
	scanner := bufio.NewScanner(file)
	defer func() {
		if err := scanner.Err(); err != nil {
			cfg.ErrorCh <- fmt.Errorf("reading %s: %w", cfg.SrcFileName, err)
		}
	}()
	for scanner.Scan() {
		// Периодически проверяем необходимость отмены
		if i%cfg.CancelCheckInterval == 0 {
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
		i++

		// Формируем пакет добавлением строк
		s := strings.TrimSpace(scanner.Text())
		if s == "" {
			continue
		}
		batch = append(batch, s)

		// По достижении заданного размера отправляем задание
		if len(batch) == cfg.BatchSize {
			if sendBatch(batch) {
				// После отправки формируем новый пакет
				batch = make([]string, 0, cfg.BatchSize)
			}
		}
	}
	// Отправляем оставшийся неполный пакет перед выходом
	if len(batch) > 0 {
		sendBatch(batch)
	}
}

// worker читает слова из cfg.JobCh, вычисляет MD5-хеш и сравнивает его с cfg.TargetHash;
// при совпадении помещает слово в cfg.ResultCh и завершается;
// при отсутствии совпадений завершается
func worker(ctx context.Context, cfg *PipelineConfig) {
	for {
		select {
		case jobs, ok := <-cfg.JobCh:
			if !ok {
				// Возврат при чтении из закрытого канала, проверка факта закрытия по второму значению (ok)
				return
			}
			for i, job := range jobs {
				if i%cfg.CancelCheckInterval == 0 {
					// Неблокирующая проверка отмены контекста: если отменен, немедленно выходим;
					// в противном случае продолжаем
					select {
					case <-ctx.Done():
						return
					default:
					}
				}
				// Увеличиваем счётчик попыток сравнения для каждого сравнения
				cfg.Counter.Add(1)
				if md5.Sum([]byte(job)) == *cfg.TargetHash {
					// Одновременно ждем отмены контекста или записи в канал;
					// после выполнения первого из этих действий переходим к return
					select {
					case cfg.ResultCh <- job:
					case <-ctx.Done():
					}
					return
				}
			}
		case <-ctx.Done():
			//
			return
		}
	}
}

// collect ожидает чтения либо закрытия cfg.ResultCh, выводит значение в консоль и завершается
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
	f, err := os.Create("cpu.prof")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		panic(err)
	}
	// Между вызовами Start и Stop будут производиться замеры
	defer pprof.StopCPUProfile()

	const srcFileName = "rockyou.txt"
	var maxWorkers = runtime.GOMAXPROCS(0) // ставим лимит не выше числа доступных ядер
	var count atomic.Uint64

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
	_, err = hex.Decode(hashBytes[:], []byte(hashStr))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	// Канал с заданиями
	jobCh := make(chan []string, maxWorkers)
	// Канал с результатами
	resultCh := make(chan string)
	// Канал с ошибками
	errCh := make(chan error, 3)

	// Флаг ошибки обработки
	var hadError atomic.Bool

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config := &PipelineConfig{
		JobCh:               jobCh,
		ResultCh:            resultCh,
		ErrorCh:             errCh,
		SrcFileName:         srcFileName,
		TargetHash:          &hashBytes,
		Counter:             &count,
		CancelCheckInterval: 64,
		BatchSize:           1024,
	}

	// Отдельная горутина для вывода попыток через равные промежутки времени
	go func() {
		ticker := time.NewTicker(time.Millisecond * 250)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				fmt.Printf("\r%d", count.Load())
			case <-ctx.Done():
				return
			}
		}
	}()

	// Сбор ошибок
	var errWg sync.WaitGroup
	errWg.Go(func() {
		for err := range errCh {
			if err != nil {
				fmt.Fprintf(os.Stderr, "error %v\n", err)
				hadError.Store(true)
			}
		}
	})

	// Генерируем задания
	// Горутина‑генератор пишет задания в канал и сама же его закрывает
	go func() {
		produce(ctx, config)
		close(jobCh)
	}()

	// Запускаем пул обработчиков, по завершении закрываем канал
	var workerWg sync.WaitGroup
	for range maxWorkers {
		workerWg.Go(func() {
			worker(ctx, config)
		})
		// Обработчики завершаются, когда задания заканчиваются и закрывается jobCh или когда отменяется контекст
	}
	go func() {
		// После завершения обработчиков писать в каналы уже некому.
		// Значит, дождавшись обработчиков, мы можем безопасно закрыть каналы.
		// Это завершит горутину - сборщик ошибок.
		workerWg.Wait()
		close(resultCh)
		close(errCh)
	}()

	// Останавливаем работу после получения результата
	collect(config)
	stop()

	// Ожидаем завершения сборки ошибок и вызываем код возврата
	errWg.Wait()
	if hadError.Load() {
		os.Exit(1)
	}
}
