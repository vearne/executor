package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/vearne/executor"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type AsyncTask struct {
	param int
}

func (m *AsyncTask) Call(ctx context.Context) *executor.GPResult {
	// do something in background
	fmt.Println("do something in background...")
	time.Sleep(3 * time.Second)
	fmt.Println("result:", m.param*m.param)
	r := executor.GPResult{}
	//r.Value = m.param * m.param
	r.Err = nil
	return &r
}

var (
	pool executor.ExecutorService
)

func init() {
	pool = executor.NewDynamicGPool(context.Background(), 5, 30,
		executor.WithDynamicTaskQueueCap(5),
		executor.WithDetectInterval(time.Second*10),
		executor.WithMeetCondNum(3),
	)
}

type TaskParam struct {
	Param int `json:"param"`
}

/*
	curl -XPOST http://localhost:8080/api/task -d '{"param":12}'
	curl -XPOST http://localhost:8080/api/task -d '{"param":10}'
*/
func main() {
	r := gin.Default()
	r.POST("/api/task", func(c *gin.Context) {
		p := TaskParam{}
		c.BindJSON(&p)
		task := AsyncTask{param: p.Param}
		_, err := pool.Submit(&task)
		if err != nil {
			log.Println("error", err)
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "submit success",
		})
	})

	// 启动web服务，监听9090端口
	go func() { log.Fatal(r.Run(":8080")) }()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	<-ch
	//pool.Cancel()
	pool.Shutdown()
	pool.WaitTerminate()
}
