package main

import (
	"fmt"

	"github.com/imattman/simple-task-worker/work"
)

type echoFactory struct {
	count int
}

func (f *echoFactory) Make(line string) work.Task {
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
func (t *echoTask) Print() {
	fmt.Printf("%d: %s\n", t.id, t.line)
}

func main() {
	work.Run(&echoFactory{}, 10)
}
