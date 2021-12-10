package pool

import (
	"sync/atomic"
	"time"
)

// 传参和返回值
type Callable func(...interface{}) interface{}

type Runnable struct {
	Arg []interface{}
	F   Callable
}

func NewFunc(f Callable, arg ...interface{}) *Runnable {
	return &Runnable{
		Arg: arg,
		F:   f,
	}
}

func (r *Runnable) Invoke() interface{} {
	return r.F(r.Arg...)
}

type IWorker interface {
	Get() interface{}
}

type Worker struct {
	f    *Runnable // 实际执行任务
	pool *Pool     // 绑定的池
}

type ResultWorker struct {
	*Worker
	timeout  time.Duration
	retChan  chan interface{}
	doneChan chan struct{}
}

func NewWorker(f Callable, arg ...interface{}) *Worker {
	return &Worker{f: NewFunc(f, arg...)}
}

func NewResultWorker(f Callable, timeout time.Duration, arg ...interface{}) *ResultWorker {
	rw := &ResultWorker{
		Worker:   NewWorker(f, arg),
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
