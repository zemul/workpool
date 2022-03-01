package workpool

import "context"

type TaskResult struct {
	Err    error
	Result interface{}
}

type Task struct {
	ctx     context.Context
	execute func(ctx context.Context) (interface{}, error)
	result  chan *TaskResult
}

func NewTask(ctx context.Context, execute func(ctx context.Context) (interface{}, error)) *Task {
	return &Task{
		ctx:     ctx,
		execute: execute,
		result:  make(chan *TaskResult, 1),
	}
}

func (t *Task) Execute() {
	var result interface{}
	var err error
	if t.execute != nil {
		result, err = t.execute(t.ctx)
	}
	t.result <- &TaskResult{Result: result, Err: err}
}

func (t *Task) Result() <-chan *TaskResult{
	return t.result
}
