package main

import (
	"fmt"

	"github.com/imattman/simple-task-worker/lib/worker"
)

type captureFactory struct {
	count int
}

func (f *captureFactory) Make(line string) task.Task {
	f.count++
	return &captureTask{line, f.count}
}

type captureTask struct {
	line string
	id   int
}

func (t *captureTask) Process() {
	// nothing
}
func (t *captureTask) Print() {
	fmt.Printf("%d: %s\n", t.id, t.line)
}

func main() {
	task.Run(&captureFactory{}, 10)
}
