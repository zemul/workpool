# workpool
简单的golang worker pool实现
- 并发限制goroutine池
- tryDo，工作池满，不等待立即返回
- 同步等待任务执行结果
```go
func main() {
	p := workpool.NewPool(context.Background(), workpool.Option{WorkerNumber: 3,TaskQueueSize: 100})
	p.Start()
	defer p.Stop()
	tasks := make([]*workpool.Task, 0)
	for i := 0; i <= 100; i++ {
		task := genTask(i)
		p.Do(task)
		tasks = append(tasks, task)
	}
	for i := range tasks {
		v := <-tasks[i].Result()
		fmt.Println(v.Result)
	}
}

func genTask(i int) *workpool.Task {
	return workpool.NewTask(context.Background(), func(ctx context.Context) (interface{}, error) {
		time.Sleep(time.Second)
		return i*i, nil
	})
}
```