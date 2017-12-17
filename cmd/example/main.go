package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"time"

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

func (t *echoTask) Process() {
	// nothing
}
func (t *echoTask) Result() (string, bool) {
	return fmt.Sprintf("%d: %s", t.id, t.line), true
}

func main() {
	var max time.Duration

	flag.DurationVar(&max, "t", 3*time.Second, "Timeout for total execution")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), max)
	defer cancel()

	scan := bufio.NewScanner(os.Stdin)
	scan.Split(bufio.ScanWords)

	task.Run(ctx, &echoFactory{}, 10,
		task.WithArgs(flag.Args()),
		task.WithScanner(scan))
}
