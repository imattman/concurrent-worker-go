package task

import (
	"bufio"
	"log"
	"os"
	"sync"
)

// Task represents a unit of work that knows how to process and print itself
type Task interface {
	Process()
	Print()
}

// Factory is responsible for transforming a line of input into a Task
type Factory interface {
	Make(line string) Task
}

// Run creates Tasks using the supplied factory and hands them to workers to
// be processed.
func Run(f Factory, numWorkers int) {
	var wg sync.WaitGroup

	unprocessed := make(chan Task)

	wg.Add(1)
	go func() {
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			unprocessed <- f.Make(s.Text())
		}
		if s.Err() != nil {
			log.Fatalf("Error reading STDIN: %s", s.Err())
		}
		close(unprocessed)
		wg.Done()
	}()

	processed := make(chan Task)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			for t := range unprocessed {
				t.Process()
				processed <- t
			}
			wg.Done()
		}()
	}

	// wait for all workers to complete
	go func() {
		wg.Wait()
		close(processed)
	}()

	for t := range processed {
		t.Print()
	}
}
