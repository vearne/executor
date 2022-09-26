package main

import (
	"context"
	"fmt"
	"github.com/vearne/executor"
	"time"
)

type MyCallable struct {
	param int
}

func (m *MyCallable) Call(ctx context.Context) *executor.GPResult {
	time.Sleep(1 * time.Second)
	r := executor.GPResult{}
	r.Value = m.param * m.param
	r.Err = nil
	return &r
}

func main() {
	//pool := executor.NewSingleGPool(context.Background())
	/*
	   options:
	   executor.WithTaskQueueCap() : set capacity of task queue
	*/
	pool := executor.NewSingleGPool(context.Background(), executor.WithTaskQueueCap(10))
	futureList := make([]executor.Future, 0)
	var f executor.Future
	var err error

	for i := 0; i < 50; i++ {
		task := &MyCallable{param: i}
		f, err = pool.Submit(task)
		if err == nil {
			fmt.Println("add task", i)
			futureList = append(futureList, f)
		}
	}

	pool.Shutdown()
	var result *executor.GPResult
	for _, f := range futureList {
		result = f.Get()
		fmt.Println(result.Err, result.Value)
	}
	pool.WaitTerminate()
}
