package task

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"sync"
)

// Task represents a unit of work that knows how to process and print itself
type Task interface {
	Process()
	Result() (result string, success bool)
}

// Factory is responsible for transforming a line of input into a Task
type Factory interface {
	Make(line string) Task
}

// Config holds configuration options for controlling Run behavior.
type Config struct {
	args    []string
	scanner *bufio.Scanner
}

func defaultConfig() *Config {
	return &Config{}
}

func WithArgs(args []string) func(*Config) {
	return func(cfg *Config) {
		cfg.args = args
	}
}

func WithScanner(s *bufio.Scanner) func(*Config) {
	return func(cfg *Config) {
		cfg.scanner = s
	}
}

// Run creates Tasks using the supplied factory and hands them to workers for processing.
func Run(ctx context.Context, f Factory, numWorkers int, options ...func(*Config)) {
	cfg := defaultConfig()
	for _, opt := range options {
		opt(cfg)
	}

	unprocessed := make(chan Task)
	processed := make(chan Task)

	var wg sync.WaitGroup

	// process commandline args / stdin
	wg.Add(1)
	go func() {
		defer func() {
			close(unprocessed)
			wg.Done()
		}()

		// consume args if any present
		if len(cfg.args) > 0 {
			for _, v := range cfg.args {
				unprocessed <- f.Make(v)
			}
			return
		}

		s := cfg.scanner
		if s == nil {
			s = bufio.NewScanner(os.Stdin)
			s.Split(bufio.ScanLines)
		}

		for s.Scan() {
			unprocessed <- f.Make(s.Text())
		}
		if s.Err() != nil {
			log.Fatalf("Error reading stdin: %s", s.Err())
		}
	}()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range unprocessed {
				t.Process()
				processed <- t
			}
		}()
	}

	// wait for all workers to complete
	go func() {
		wg.Wait()
		close(processed)
	}()

	complete := make(chan bool, 1)
	go func() {
		for t := range processed {
			res, _ := t.Result()
			fmt.Println(res)
		}
		complete <- true
	}()

	select {
	case <-complete:
	case <-ctx.Done():
		fmt.Fprintf(os.Stderr, "Job timed out: %s\n", ctx.Err())
		os.Exit(1)
	}
}
