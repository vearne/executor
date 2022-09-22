# executor
[![golang-ci](https://github.com/vearne/executor/actions/workflows/golang-ci.yml/badge.svg)](https://github.com/vearne/executor/actions/workflows/golang-ci.yml)

goroutine pool

* [中文 README](https://github.com/vearne/executor/blob/master/README_zh.md)

## Feature
* supports cancel s single task or cancel all tasks in the goroutine pool.
* Multiple types of goroutine pools can be created(SingleGPool|FixedGPool|DynamicGPool).

## Multiple types of goroutine pools
|||
|:---|:---|:---|

## debug
optional value: debug | info | warn | error
```
export SIMPLE_LOG_LEVEL=debug
```

## Inspired by Java Executors
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
	time.Sleep(1 * time.Second)
	r := executor.GPResult{}
	r.Value = m.param * m.param
	return &r
}

func main() {
	pool := executor.NewFixedGPool(context.Background(), 10)
	futureList := make([]executor.Future, 0)
	var f executor.Future
	for i := 0; i < 10; i++ {
		task := &MyCallable{param: i}
		f = pool.Submit(task)
		futureList = append(futureList, f)
	}
	pool.Shutdown() // Prohibit submission of new tasks
	//pool.Submit(&MyCallable{param: 20})
	var result *executor.GPResult
	for _, f := range futureList {
		result = f.Get()
		fmt.Println(result.Err, result.Value)
	}
	pool.WaitTerminate()
}
```