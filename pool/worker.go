package pool

import (
	"sync/atomic"
	"time"
)

type IWorker interface {
	Get() interface{}
}

type Worker struct {
	f    func() interface{}
	pool *Pool
}

type ResultWorker struct {
	*Worker
	timeout  time.Duration
	retChan  chan interface{}
	doneChan chan struct{}
}

func NewWorker(f func() interface{}) *Worker {
	return &Worker{f: f}
}

func NewResultWorker(f func() interface{}, timeout time.Duration) *ResultWorker {
	rw := &ResultWorker{
		Worker:   NewWorker(f),
		retChan:  make(chan interface{}, 1),
		doneChan: make(chan struct{}, 1),
		timeout:  timeout,
	}
	return rw
}

func (w *Worker) Get() interface{} {
	if atomic.LoadInt32(&w.pool.shutdown) == 1 {
		panic(ErrPoolShutdown)
	}
	return nil
}

// Get 获取worker执行返回的结果
func (w *ResultWorker) Get() interface{} {
	// Pool was shutdown
	if atomic.LoadInt32(&w.pool.shutdown) == 1 {
		panic(ErrPoolShutdown)
	}
	select {
	case <-w.doneChan: // 阻塞等待通知
		select {
		case ret := <-w.retChan: // 取出返回值
			return ret
		default:
		}
	}
	return nil
}
