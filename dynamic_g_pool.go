package executor

import (
	"context"
	slog "github.com/vearne/simplelog"
	"sync"
	"sync/atomic"
	"time"
)

// The number of worker in DynamicGPool changes dynamically. The minimum is Min and the maximum is Max.
// Expansion rules: If the TaskChan is full, try to add workers to execute the task.
// Shrinking rules:
//     Condition: If the number of workers in a busy state is less than 1/4 of the total number of workers,
//     try to reduce the number of workers by 1/2. Execute meetCondNum consecutive checks,
//     with detectInterval every time, and perform shrinking if the conditions are met each time.

type DynamicGPoolOption struct {
	taskQueueCap int
	// interval between checks
	detectInterval time.Duration
	// the number of times the shrinkage is performed to meet the conditions
	meetCondNum int
}

type dynamicOption func(*DynamicGPoolOption)

// Optional parameters
func WithDynamicTaskQueueCap(taskQueueCap int) dynamicOption {
	return func(t *DynamicGPoolOption) {
		t.taskQueueCap = taskQueueCap
	}
}

func WithDetectInterval(detectInterval time.Duration) dynamicOption {
	return func(t *DynamicGPoolOption) {
		t.detectInterval = detectInterval
	}
}

func WithMeetCondNum(meetCondNum int) dynamicOption {
	return func(t *DynamicGPoolOption) {
		t.meetCondNum = meetCondNum
	}
}

type DynamicGPool struct {
	wg sync.WaitGroup

	min int32
	max int32

	currGCount int32
	workerList []*Worker
	// rwMutex to protect workerList
	rwMutex sync.RWMutex

	// task queue
	TaskChan   chan *FutureTask
	isShutdown *AtomicBool
	// Context
	ctx    context.Context
	cancel context.CancelFunc
}

func NewDynamicGPool(ctx context.Context, min int, max int, opts ...dynamicOption) ExecutorService {
	// check params
	if min <= 0 {
		min = 1
	}

	if min > max {
		panic("min must be less than or equal to max")
	}

	defaultOpts := &DynamicGPoolOption{
		taskQueueCap:   SIZE,
		detectInterval: time.Minute,
		meetCondNum:    3,
	}
	// Loop through each option
	for _, opt := range opts {
		// Call the option giving the instantiated
		opt(defaultOpts)
	}

	pool := DynamicGPool{}
	pool.ctx, pool.cancel = context.WithCancel(ctx)
	pool.isShutdown = NewAtomicBool(false)
	pool.TaskChan = make(chan *FutureTask, defaultOpts.taskQueueCap)
	pool.min = int32(min)
	pool.max = int32(max)

	pool.rwMutex.Lock()
	for i := 0; i < min; i++ {
		w := NewWorker(&pool)
		pool.workerList = append(pool.workerList, w)
		go w.Start()
	}
	atomic.StoreInt32(&pool.currGCount, pool.min)
	pool.rwMutex.Unlock()

	return &pool
}

func (p *DynamicGPool) Shutdown() {
	close(p.TaskChan)
	p.isShutdown.Set(true)
}

// When submitting tasks, blocking may occur
func (p *DynamicGPool) Submit(task Callable) Future {
	p.wg.Add(1)

	t := NewFutureTask(p.ctx, task)
	select {
	case p.TaskChan <- t:
		// push into TaskChan
		slog.Debug("add task to TaskChan")
	default:
		// If the TaskChan is full, try to add workers to execute the task
		curr := atomic.LoadInt32(&p.currGCount)
		if curr < p.max {
			newValue := atomic.AddInt32(&p.currGCount, 1)
			if newValue <= p.max {
				// create new worker
				p.rwMutex.Lock()

				w := NewWorker(p)
				p.workerList = append(p.workerList, w)
				p.rwMutex.Unlock()
				go w.Start()
			} else {
				// rollback
				atomic.AddInt32(&p.currGCount, -1)
			}
		}
		// blocking may occur
		p.TaskChan <- t
	}
	return t
}

func (p *DynamicGPool) IsShutdown() bool {
	return p.isShutdown.IsTrue()
}

func (p *DynamicGPool) CurrGCount() int {
	return int(atomic.LoadInt32(&p.currGCount))
}

func (p *DynamicGPool) WaitTerminate() {
	if !p.IsShutdown() {
		panic("pool must shutdown first!")
	}
	p.wg.Wait()
}

func (p *DynamicGPool) TaskQueueCap() int {
	return cap(p.TaskChan)
}

func (p *DynamicGPool) TaskQueueLength() int {
	return len(p.TaskChan)
}

func (p *DynamicGPool) Cancel() bool {
	p.cancel()
	return true
}

type ShrinkWorker struct {
	RunningFlag *AtomicBool
	ExitedFlag  chan struct{}
	ExitChan    chan struct{}
	pool        *DynamicGPool
	// ----- shrink related --------
	// interval between checks
	detectInterval time.Duration
	// the number of times the shrinkage is performed to meet the conditions
	meetCondNum int
	// The current number of times the condition is met
	currMeetCond int
}

func NewShrinkWorker(pool *DynamicGPool, interval time.Duration, meetCondNum int) *ShrinkWorker {
	worker := ShrinkWorker{}
	worker.RunningFlag = NewAtomicBool(true)
	worker.ExitedFlag = make(chan struct{})
	worker.ExitChan = make(chan struct{})
	worker.pool = pool
	worker.detectInterval = interval
	worker.meetCondNum = meetCondNum
	worker.currMeetCond = 0
	return &worker
}

// Shrinking rules:
//     Condition: If the number of workers in a busy state is less than 1/4 of the total number of workers,
//     try to reduce the number of workers by 1/2. Execute meetCondNum consecutive checks,
//     with detectInterval every time, and perform shrinking if the conditions are met each time.
func (w *ShrinkWorker) Start() {
	ticker := time.NewTicker(w.detectInterval)
	for w.RunningFlag.IsTrue() {
		select {
		case <-ticker.C:
			slog.Debug("ShrinkWorker check")

			busyCount := 0
			w.pool.rwMutex.RLock()
			for _, worker := range w.pool.workerList {
				if worker.IsBusy() {
					busyCount++
				}
			}
			w.pool.rwMutex.RUnlock()

			// < 1/4
			if float64(busyCount)/float64(atomic.LoadInt32(&w.pool.currGCount)) < 0.25 {
				w.currMeetCond++
			}
			if w.currMeetCond >= 3 { // execute shrink
				w.currMeetCond = 0

				w.pool.rwMutex.Lock()
				// Put busy workers at the head of the array and idle workers at the end
				reorganize(w.pool.workerList)
				currGCount := atomic.LoadInt32(&w.pool.currGCount)
				for i := currGCount - 1; i > currGCount/2; i-- {
					w.pool.workerList[i].Stop()
					atomic.AddInt32(&w.pool.currGCount, -1)
				}
				w.pool.workerList = w.pool.workerList[0 : currGCount-currGCount/2]
				w.pool.rwMutex.Unlock()
			}
		case <-w.ExitChan:
			slog.Debug("ShrinkWorker exiting.")
		}
	}
	close(w.ExitedFlag)
}

// Put busy workers at the head of the array and idle workers at the end
// [busy, busy, idle, idle, idle]
func reorganize(list []*Worker) {
	N := len(list)
	i, j := 0, len(list)-1
	for i < j {
		for i < N && list[i].IsBusy() {
			i++
		}
		for j >= 0 && !list[j].IsBusy() {
			j--
		}
		if i < j {
			list[i], list[j] = list[j], list[i]
		}
	}
}

func (worker *ShrinkWorker) Stop() {
	worker.RunningFlag.Set(false)
	close(worker.ExitChan)

	<-worker.ExitedFlag
}

type Worker struct {
	RunningFlag *AtomicBool
	ExitedFlag  chan struct{}
	ExitChan    chan struct{}
	pool        *DynamicGPool

	// is worker busy?
	busyFlag *AtomicBool
}

func NewWorker(pool *DynamicGPool) *Worker {
	worker := Worker{}
	worker.busyFlag = NewAtomicBool(false)
	worker.RunningFlag = NewAtomicBool(true)
	worker.ExitedFlag = make(chan struct{})
	worker.ExitChan = make(chan struct{})
	worker.pool = pool
	return &worker
}

func (worker *Worker) IsBusy() bool {
	return worker.busyFlag.IsTrue()
}

func (worker *Worker) Start() {
	worker.execute()
}

func (worker *Worker) execute() {
	for worker.RunningFlag.IsTrue() {
		select {
		case task := <-worker.pool.TaskChan:
			worker.busyFlag.Set(true)
			task.run()
			worker.busyFlag.Set(false)

			worker.pool.wg.Done()
		case <-worker.ExitChan:
			// exit
		}
	}
	close(worker.ExitedFlag)
}

func (worker *Worker) Stop() {
	worker.RunningFlag.Set(false)
	close(worker.ExitChan)

	<-worker.ExitedFlag
}
