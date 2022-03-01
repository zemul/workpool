package workpool

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
)

type Option struct {
	WorkerNumber int
	TaskQueueSize int
}

type Pool struct {
	ctx    context.Context
	cancal context.CancelFunc

	o Option

	wg        sync.WaitGroup
	taskQueue chan *Task

	state int32 // 0 stop, 1 start
}

func (o *Option) normalize() {
	if o.WorkerNumber <= 0 {
		o.WorkerNumber = runtime.NumCPU()
	}
	if o.TaskQueueSize <= 0{
		o.TaskQueueSize = runtime.NumCPU()
	}
}

func NewPool(ctx context.Context, o Option) (p *Pool) {
	if ctx == nil {
		ctx = context.Background()
	}
	o.normalize()

	p = &Pool{
		o:         o,
		taskQueue: make(chan *Task, o.TaskQueueSize),
	}
	p.ctx, p.cancal = context.WithCancel(ctx)
	return p
}

func (p *Pool) Start() {
	if atomic.CompareAndSwapInt32(&p.state, 0, 1) {
		p.wg.Add(p.o.WorkerNumber)
		for i := 0; i < p.o.WorkerNumber; i++ {
			go p.worker()
		}

	}
}

func (p *Pool) Stop() {
	if atomic.CompareAndSwapInt32(&p.state, 1, 0) {
		// cancel context
		p.cancal()

		// wait workers exit
		close(p.taskQueue)
		p.wg.Wait()
	}
}

func (p *Pool) Execute(exec func(ctx context.Context) (interface{}, error)) (t *Task) {
	return p.ExecuteWithCtx(p.ctx, exec)
}

func (p *Pool) ExecuteWithCtx(ctx context.Context, exec func(ctx context.Context) (interface{}, error)) (t *Task) {
	if ctx == nil {
		ctx = context.Background()
	}
	t = NewTask(ctx, exec)
	p.Do(t)
	return
}

func (p *Pool) TryExecute(exec func(ctx context.Context) (interface{}, error)) (t *Task,addQueue bool) {
	return p.TryExecuteWithCtx(p.ctx, exec)
}

func (p *Pool) TryExecuteWithCtx(ctx context.Context, exec func(ctx context.Context) (interface{}, error)) (t *Task, addQueue bool) {
	if ctx == nil {
		ctx = context.Background()
	}
	t = NewTask(ctx, exec)
	addQueue = p.TryDo(t)
	return
}

func (p *Pool) Do(task *Task) {
	if task != nil {
		select {
		case <-p.ctx.Done():
		case p.taskQueue <- task:
		}
	}
}

// TryDo Try to write, if  taskQueue is full, immediately return
func (p *Pool) TryDo(task *Task) (addTaskSuccess bool) {
	if task != nil {
		select {
		case p.taskQueue <- task:
			addTaskSuccess = true
		case <-p.ctx.Done():
		default:

		}
	}
	return
}

func (p *Pool) worker() {
	for task := range p.taskQueue {
		task.Execute()
	}
	p.wg.Done()
}
