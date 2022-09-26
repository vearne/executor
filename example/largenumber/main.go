package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/vearne/executor"
	"io"
	"log"
	"os"
	"strconv"
)

/*
	When there are many tasks to be executed, a large number of task parameters or
	task execution results will be accumulated in the memory, resulting in huge memory overhead.
	If the task parameters are read from a database such as MySQL, be careful
	about the timeout problem of the connection with MySQL.
*/

const (
	dataFilePath = "/tmp/data.csv"
)

type MyCallable struct {
	param int
}

func (m *MyCallable) Call(ctx context.Context) *executor.GPResult {
	//time.Sleep(100 * time.Millisecond)
	r := executor.GPResult{}
	r.Value = m.param * m.param
	r.Err = nil
	return &r
}

func main() {
	// Generate data files
	genDataCSV()

	pool := executor.NewFixedGPool(context.Background(), 50)
	futureCh := make(chan executor.Future, 10)

	// Child goroutines act as producers of tasks
	go func() {
		var (
			f   executor.Future
			err error
		)

		// open file
		file, err := os.Open(dataFilePath)
		if err != nil {
			log.Fatal(err)
		}
		// remember to close the file at the end of the program
		defer file.Close()
		// read csv values using csv.Reader
		csvReader := csv.NewReader(file)

		for {
			rec, err := csvReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}

			param, _ := strconv.Atoi(rec[0])
			task := &MyCallable{param: param}
			f, err = pool.Submit(task)
			if err != nil {
				log.Fatal(err)
			} else {
				futureCh <- f
			}
		}
		close(futureCh)
	}()

	// Goroutine main act as a consumer
	for f := range futureCh {
		result := f.Get()
		fmt.Println(result.Err, result.Value)
	}
	pool.Shutdown()
	pool.WaitTerminate()
}

func genDataCSV() {
	file, err := os.Create(dataFilePath)
	defer file.Close()
	if err != nil {
		log.Fatalln("failed to open file", err)
	}
	w := csv.NewWriter(file)
	defer w.Flush()
	// Using Write
	for i := 0; i < 100001; i++ {
		row := []string{strconv.Itoa(i)}
		if err := w.Write(row); err != nil {
			log.Fatalln("error writing record to file", err)
		}
	}
}
