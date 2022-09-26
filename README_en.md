# executor
[![golang-ci](https://github.com/vearne/executor/actions/workflows/golang-ci.yml/badge.svg)](https://github.com/vearne/executor/actions/workflows/golang-ci.yml)

goroutine pool

* [中文 README](https://github.com/vearne/executor/blob/master/README_zh.md)

## 1. Feature
* supports cancel s single task or cancel all tasks in the goroutine pool.
```
Future.Cancel()
```
```
ExecutorService.Cancel()
```
You can also use context.Context to cancel task or pool.

* Multiple types of goroutine pools can be created(SingleGPool|FixedGPool|DynamicGPool).

## 2. Multiple types of goroutine pools
|category| explain                                                                      | remark                                                           |
|:---|:------------------------------------------------------------------------|:-----------------------------------------------------------------|
|SingleGPool| A single worker goroutine pool                                          |                                                                  |
|FixedGPool| Fixed number of worker goroutine pools                                  |                                                                  |
|DynamicGPool| A goroutine pool where the number of workers can be dynamically changed | min: Minimum number of workers<br/> max:Maximum number of coroutines |

### 2.1 SingleGPool
```
NewSingleGPool(ctx context.Context, opts ...option) ExecutorService
```

### 2.2 FixedGPool
```
NewFixedGPool(ctx context.Context, size int, opts ...option) ExecutorService
```
### 2.3 DynamicGPool
```
NewDynamicGPool(ctx context.Context, min int, max int, opts ...dynamicOption) ExecutorService
```

#### 2.3.1 Expansion rules
If the task queue is full, try to add workers to execute the task.

#### 2.3.2 Shrinking rules:
* `Condition`: If the number of workers in a busy state is less than 1/4 of the total number of workers, the condition is considered satisfied
* Perform `meetCondNum` consecutive checks, each with a `detectInterval` interval. If the conditions are met every time, the scaling is triggered.
* The scaling action tries to reduce the number of workers by half

* `Condition`: If the number of workers in a busy state is less than 1/4 of the total number of workers,try to reduce the number of workers by 1/2.
* Execute `meetCondNum` consecutive checks, with `detectInterval` every time, and perform shrinking if the conditions are met each time.

## 3. Notice
Since the executor uses the channel as the task queue, blocking may occur when submitting tasks.
```
Submit(task Callable) (Future, error)
```
If the goroutine pool is running in the background for a long time, we strongly recommend monitoring the usage of the task queue.
```
TaskQueueCap() int
TaskQueueLength() int
```


## 4. Example
[more examples](https://github.com/vearne/executor/tree/main/example)

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

	for i := 0; i < 100; i++ {
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
```

## 5. debug
set log level
optional value: debug | info | warn | error
```
export SIMPLE_LOG_LEVEL=debug
```

## 6. Thanks
Inspired by Java Executors
[Executors](https://docs.oracle.com/en/java/javase/11/docs/api/java.base/java/util/concurrent/Executors.html)