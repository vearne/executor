package main

import (
	"context"
	"fmt"
	"github.com/vearne/executor"
	"math/rand"
	"net/http"
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
	//pool := executor.NewDynamicGPool(context.Background(), 5, 30)
	/*
	   options:
	   executor.WithTaskQueueCap() : set capacity of task queue
	*/
	pool := executor.NewDynamicGPool(context.Background(), 3, 100,
		executor.WithDynamicTaskQueueCap(5),
		executor.WithDetectInterval(time.Second*5),
		executor.WithMeetCondNum(3),
	)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 往w里写入内容，就会在浏览器里输出
		fmt.Fprintf(w, "CurrentGCount:%v\n", pool.CurrentGCount())
	})

	// 启动web服务，监听9090端口
	go func() { http.ListenAndServe(":8080", nil) }()

	var err error
	go func() {
		running := executor.NewAtomicBool(true)
		time.AfterFunc(time.Second*10, func() {
			running.Set(false)
		})
		for running.IsTrue() {
			task := &MyCallable{param: rand.Intn(200)}
			_, err = pool.Submit(task)
			if err == nil {
				fmt.Println("add task success")
			}
		}
		time.Sleep(2 * time.Minute)
		running.Set(true)
		time.AfterFunc(time.Second*10, func() {
			running.Set(false)
		})
		for running.IsTrue() {
			task := &MyCallable{param: rand.Intn(200)}
			_, err = pool.Submit(task)
			if err == nil {
				fmt.Println("add task success")
			}
		}
	}()
	go func() {
		for {
			time.Sleep(time.Millisecond * 200)
			task := &MyCallable{param: rand.Intn(200)}
			_, err = pool.Submit(task)
			if err == nil {
				fmt.Println("add task success")
			}
		}
	}()
	time.Sleep(5 * time.Minute)
	pool.Shutdown() // Prohibit submission of new tasks
	pool.WaitTerminate()
}
