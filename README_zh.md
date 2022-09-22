# executor
[![golang-ci](https://github.com/vearne/executor/actions/workflows/golang-ci.yml/badge.svg)](https://github.com/vearne/executor/actions/workflows/golang-ci.yml)

goroutine pool

* [English README](https://github.com/vearne/executor/blob/master/README.md)

## 特性
* 支持取消单个任务，也可以取消协程池上的所有任务
* 可以创建多种不同类型的协程池（SingleGPool|FixedGPool|DynamicGPool）

## 多种类型的协程池
|类别| 说明           | 备注                     |
|:---|:-------------|:-----------------------|
|SingleGPool| 单个worker协程池      |                        |
|FixedGPool| 固定数量worker协程池      |                        |
|DynamicGPool| worker数量可以动态变化的协程池 | min: 最少协程数量; max最大协程数量 |

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

## 调试
设置日志级别
可选值: debug | info | warn | error
```
export SIMPLE_LOG_LEVEL=debug
```

## 感谢
本项目受到Java Executors的启发
[Executors](https://docs.oracle.com/en/java/javase/11/docs/api/java.base/java/util/concurrent/Executors.html)

## 示例
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