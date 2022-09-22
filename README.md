# executor
[![golang-ci](https://github.com/vearne/executor/actions/workflows/golang-ci.yml/badge.svg)](https://github.com/vearne/executor/actions/workflows/golang-ci.yml)

goroutine pool

* [中文 README](https://github.com/vearne/executor/blob/master/README_zh.md)

## Feature
* supports cancel s single task or cancel all tasks in the goroutine pool.
* Multiple types of goroutine pools can be created(SingleGPool|FixedGPool|DynamicGPool).

## Multiple types of goroutine pools
|category| explain                                                                      | remark                                                           |
|:---|:------------------------------------------------------------------------|:-----------------------------------------------------------------|
|SingleGPool| A single worker goroutine pool                                          |                                                                  |
|FixedGPool| Fixed number of worker goroutine pools                                  |                                                                  |
|DynamicGPool| A goroutine pool where the number of workers can be dynamically changed | min: Minimum number of workers; max:Maximum number of coroutines |

### SingleGPool
```
executor.NewSingleGPool(context.Background(), executor.WithTaskQueueCap(10))
```

### FixedGPool
```
executor.NewFixedGPool(context.Background(), 10, executor.WithTaskQueueCap(10))
```
### DynamicGPool
```
executor.NewDynamicGPool(context.Background(), 5, 30,
    executor.WithDynamicTaskQueueCap(5),
    executor.WithDetectInterval(time.Second*10),
    executor.WithMeetCondNum(3),
)
```

## debug
set log level
optional value: debug | info | warn | error
```
export SIMPLE_LOG_LEVEL=debug
```

## Thanks
Inspired by Java Executors
[Executors](https://docs.oracle.com/en/java/javase/11/docs/api/java.base/java/util/concurrent/Executors.html)

## Example
```
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
	time.Sleep(3 * time.Second)
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
		for i := 0; i < 1000; i++ {
			task := &MyCallable{param: i}
			f, err = pool.Submit(task)
			if err == nil {
				fmt.Println("add task", i)
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
```