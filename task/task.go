package task

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"sync"
)

// Task represents a unit of work that knows how to process itself and later yield the result.
type Task interface {
	Process()
	Result() (result string, success bool)
}

// Factory is responsible for transforming a string of input into a Task.
type Factory interface {
	Make(raw string) Task
}

// Config holds configuration options for affecting Run behavior.
type Config struct {
	args     []string
	scanner  *bufio.Scanner
	reporter func(completed Task)
}

func defaultConfig() *Config {
	return &Config{
		reporter: func(t Task) {
			r, _ := t.Result()
			fmt.Println(r)
		},
	}
}

// WithArgs specifies a slice of raw arguments as source input to a Factory.
// Behavior falls back to reading values from the Scanner if the supplied argument slice is empty.
func WithArgs(args []string) func(*Config) {
	return func(cfg *Config) {
		cfg.args = args
	}
}

// WithScanner overrides the default Scanner used for supplying values to a Factory.
// The default scanner tokenizes lines of text read from STDIN.
func WithScanner(s *bufio.Scanner) func(*Config) {
	return func(cfg *Config) {
		cfg.scanner = s
	}
}

// WithReporter overrides the reporting function applied to completed task results.
func WithReporter(f func(completed Task)) func(*Config) {
	return func(cfg *Config) {
		cfg.reporter = f
	}
}

// Run creates Tasks using the supplied Factory and forwards the tasks to a pool of workers for concurrent processing.
func Run(ctx context.Context, f Factory, numWorkers int, options ...func(*Config)) {
	cfg := defaultConfig()
	for _, opt := range options {
		opt(cfg)
	}

	// need at least one worker
	if numWorkers < 1 {
		numWorkers = 1
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

		scanner := cfg.scanner
		if scanner == nil {
			scanner = bufio.NewScanner(os.Stdin)
			scanner.Split(bufio.ScanLines)
		}

		for scanner.Scan() {
			unprocessed <- f.Make(scanner.Text())
		}
		if scanner.Err() != nil {
			log.Fatalf("Error reading stdin: %s", scanner.Err())
		}
	}()

	// do the actual processing
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
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

	reported := make(chan bool, 1)
	go func() {
		for t := range processed {
			cfg.reporter(t)
		}
		reported <- true
	}()

	// wait for either reporting to complete or work cancelled via context
	select {
	case <-reported:
	case <-ctx.Done():
		log.Fatalf("Job timed out: %s\n", ctx.Err())
	}
}
