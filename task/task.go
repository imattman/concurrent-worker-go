package task

import (
	"bufio"
	"context"
	"fmt"
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
	args       []string
	scanner    *bufio.Scanner
	reportFunc func(completed Task)
}

func defaultConfig() *Config {
	return &Config{
		reportFunc: func(t Task) {
			r, _ := t.Result()
			fmt.Println(r)
		},
	}
}

// Run creates Tasks using the supplied Factory and forwards the tasks to a pool of workers for concurrent processing.
func Run(ctx context.Context, factory Factory, numWorkers int, options ...func(*Config)) error {
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
	errors := make(chan error, 1)

	var wg sync.WaitGroup

	// tokenize commandline args / stdin to be converted to Tasks by the Factory
	wg.Add(1)
	go func() {
		defer func() {
			close(unprocessed)
			wg.Done()
		}()

		// consume args if any present
		if len(cfg.args) > 0 {
			for _, v := range cfg.args {
				unprocessed <- factory.Make(v)
			}
			return
		}

		scan := cfg.scanner
		if scan == nil {
			scan = bufio.NewScanner(os.Stdin)
			scan.Split(bufio.ScanLines)
		}

		for scan.Scan() {
			unprocessed <- factory.Make(scan.Text())
		}
		if scan.Err() != nil {
			errors <- fmt.Errorf("error reading stdin: %s", scan.Err())
			return
		}
	}()

	// do the actual Task processing
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()

			for t := range unprocessed {
				t.Process()

				select {
				case processed <- t:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// wait for all workers to complete
	go func() {
		wg.Wait()
		close(processed)
	}()

	finished := make(chan struct{}, 1)
	go func() {
		defer func() { finished <- struct{}{} }()

		for {
			select {
			case t, ok := <-processed:
				if !ok {
					return
				}
				cfg.reportFunc(t)

			case <-ctx.Done():
				return
			}
		}
	}()

	// wait for either reporting to complete or work cancelled via context
	select {
	case <-finished:
	case err := <-errors:
		return err
	case <-ctx.Done():
		return fmt.Errorf("job timed out: %s", ctx.Err())
	}

	return nil
}
