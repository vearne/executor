# executor
[![golang-ci](https://github.com/vearne/executor/actions/workflows/golang-ci.yml/badge.svg)](https://github.com/vearne/executor/actions/workflows/golang-ci.yml)

goroutine pool

* [English README](https://github.com/vearne/executor/blob/master/README_en.md)

## 1. 特性
* 支持取消单个任务，也可以取消协程池上的所有任务
```
Future.Cancel()
```
```
ExecutorService.Cancel()
```
你也可以使用context.Context取消task或者pool。

* 可以创建多种不同类型的协程池（SingleGPool|FixedGPool|DynamicGPool）

## 2. 多种类型的协程池
|类别| 说明           | 备注                     |
|:---|:-------------|:-----------------------|
|SingleGPool| 单个worker协程池      |                        |
|FixedGPool| 固定数量worker协程池      |                        |
|DynamicGPool| worker数量可以动态变化的协程池 | min:最少协程数量<br/> max:最大协程数量 |

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
#### 2.3.1 扩容规则
提交任务时，如果任务队列已经满了，则尝试增加worker去执行任务。

#### 2.3.2 缩容规则:
* `条件`: 如果处于忙碌状态的worker少于worker总数的1/4，则认为满足条件
* 执行`meetCondNum`次连续检测，每次间隔`detectInterval`。如果每次都满足条件，触发缩容。
* 缩容动作尝试减少一半的worker

## 3. 注意
由于executor使用了channel作为作为任务队列，所以提交任务时，可能会发生阻塞。
```
Submit(task Callable) (Future, error)
```
如果协程池长期在后台执行，我们强烈建议监控任务队列的使用情况。
```
TaskQueueCap() int
TaskQueueLength() int
```

## 4. 示例
[更多示例](https://github.com/vearne/executor/tree/main/example)
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
	# 禁止提交新的任务(之前已经提交的任务不受影响)
	pool.Shutdown()
	var result *executor.GPResult
	for _, f := range futureList {
		result = f.Get()
		fmt.Println(result.Err, result.Value)
	}
	# 等待所有任务执行完毕
	pool.WaitTerminate()
}
```

## 5. 调试
设置日志级别
可选值: debug | info | warn | error
```
export SIMPLE_LOG_LEVEL=debug
```

## 6. 感谢
本项目受到Java Executors的启发
[Executors](https://docs.oracle.com/en/java/javase/11/docs/api/java.base/java/util/concurrent/Executors.html)