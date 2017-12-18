package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"log"

	"github.com/imattman/concurrent-worker-go/task"
)

type echoFactory struct {
	count int
}

func (f *echoFactory) Make(line string) task.Task {
	f.count++
	return &echoTask{line, f.count}
}

type echoTask struct {
	line string
	id   int
}

func (t *echoTask) Process() { /* nothing */ }

func (t *echoTask) Result() (string, bool) {
	return fmt.Sprintf("%d: %s", t.id, t.line), true
}

func main() {
	var max time.Duration
	var nWorkers int

	flag.DurationVar(&max, "t", 3*time.Second, "Timeout for total execution")
	flag.IntVar(&nWorkers, "n", 1, "Number of concurrent workers")
	flag.Parse()

	if nWorkers < 1 {
		fmt.Fprintf(os.Stderr, "Value for worker count must be greater than zero: %d\n\n", nWorkers)
		flag.Usage()
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), max)
	defer cancel()

	// tokenize stdin, split on words not lines
	scan := bufio.NewScanner(os.Stdin)
	scan.Split(bufio.ScanWords)

	err := task.Run(ctx, &echoFactory{}, nWorkers,
		task.WithArgs(flag.Args()),
		task.WithScanner(scan))

	if err != nil {
		log.Fatalln(err)
	}
}
