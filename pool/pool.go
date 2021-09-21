package pool

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

var (
	ErrPoolShutdown = errors.New("goroutine pool was shutdown")
	ErrWorkerIsNil  = errors.New("worker is nil")
)

type Pool struct {
	shutdown     int32
	currentWorks int32
	maxWorkers   int
	workerChan   chan IWorker
	quitChan     chan struct{}
	doneChan     chan struct{}
}

func NewPool(workers int) *Pool {
	p := &Pool{
		maxWorkers: workers,
		workerChan: make(chan IWorker),
		quitChan:   make(chan struct{}),
		doneChan:   make(chan struct{}),
	}
	p.start()
	return p
}

func (p *Pool) start() {
	for i := 0; i < p.maxWorkers; i++ {
		go p.run()
	}
}

func (p *Pool) run() {
	for {
		select {
		case worker := <-p.workerChan:
			if w, ok := worker.(*ResultWorker); ok {
				var ctx context.Context
				var cancel context.CancelFunc

				if w.timeout <= 0 {
					// 直到调用cancel
					ctx, cancel = context.WithCancel(context.Background())
				} else {
					// 超时或者调用cancel
					ctx, cancel = context.WithTimeout(context.Background(), w.timeout)
				}

				go func(c context.Context) {
					select {
					case w.retChan <- w.f():
						// 当 w.f() 能够在时间 t.timeout 内完成，将结果存放到 retChan 中
						cancel()
					}
				}(ctx)

				select {
				case <-ctx.Done():
					// 通知Worker可以Get返回结果
					// 1、w.f() 未超时，则 w.retChan 保存结果
					// 2、w.f() 超时
					w.doneChan <- struct{}{}
				}
			} else if w, ok := worker.(*Worker); ok {
				// 无返回值
				_ = w.f()
			}
			// 子协程执行完任务，通知用户可以继续submit worker
			p.doneChan <- struct{}{}
		case <-p.quitChan: // 退出子协程
			return
		}
	}
}

func (p *Pool) Submit(f func() interface{}) *Worker {
	if p.IsShutdown() {
		panic(ErrPoolShutdown)
	}
	worker := NewWorker(f)
	worker.pool = p
	p.submit(worker)
	return worker
}

func (p *Pool) SubmitResult(f func() interface{}, timeout time.Duration) *ResultWorker {
	if p.IsShutdown() {
		panic(ErrPoolShutdown)
	}
	worker := NewResultWorker(f, timeout)
	worker.pool = p
	p.submit(worker)
	return worker
}

func (p *Pool) submit(worker IWorker) {
	if worker == nil {
		panic(ErrWorkerIsNil)
	}

	for {
		select {
		case p.workerChan <- worker: // 将worker传递给某个子协程
			atomic.AddInt32(&p.currentWorks, 1)
			return
		case <-p.doneChan: // 某个子协程执行完后通知用户可以继续submit
			atomic.AddInt32(&p.currentWorks, -1)
		}
	}
}

func (p *Pool) IsShutdown() bool {
	return atomic.LoadInt32(&p.shutdown) == 1
}

func (p *Pool) Shutdown() {
	// 等待当前所有worker
	count := atomic.LoadInt32(&p.currentWorks)
	for i := 0; i < int(count); i++ {
		<-p.doneChan
	}
	// 退出所有子协程
	for i := 0; i < p.maxWorkers; i++ {
		p.quitChan <- struct{}{}
	}
	atomic.StoreInt32(&p.currentWorks, 0)
	atomic.StoreInt32(&p.shutdown, 1)
	close(p.workerChan)
	close(p.quitChan)
	close(p.doneChan)
}
