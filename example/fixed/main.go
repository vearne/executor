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
	//pool := executor.NewFixedGPool(context.Background(), 10)
	/*
	   options:
	   executor.WithTaskQueueCap() : set capacity of task queue
	*/
	pool := executor.NewFixedGPool(context.Background(), 10, executor.WithTaskQueueCap(10))
	futureList := make([]executor.Future, 0)
	var f executor.Future
	var err error
	go func() {
		for i := 0; i < 100; i++ {
			task := &MyCallable{param: i}
			f, err = pool.Submit(task)
			if err == nil {
				futureList = append(futureList, f)
			}
		}
	}()

	time.Sleep(10 * time.Second)
	pool.Shutdown()
	var result *executor.GPResult
	for _, f := range futureList {
		result = f.Get()
		fmt.Println(result.Err, result.Value)
	}
	pool.WaitTerminate()
}
