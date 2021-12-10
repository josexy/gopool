package main

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/josexy/gopool/pool"
)

var cnt int64

const SIZE = 10000000

func main() {
	p := pool.NewPool(runtime.NumCPU())
	defer func() { fmt.Println(cnt) }()
	defer p.Shutdown()

	arg := struct {
		name string
		age  int
	}{
		"mike", 12,
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		p.Submit(func(i ...interface{}) interface{} {
			fmt.Println("==> ", i)
			for j := 0; j < SIZE; j++ {
				atomic.AddInt64(&cnt, 1)
			}
			return nil
		}, arg, "hello world")
	}

	var rws []*pool.ResultWorker
	for i := 0; i < runtime.NumCPU(); i++ {
		resW := p.SubmitResult(func(i ...interface{}) interface{} {
			fmt.Println("==> ", i)
			time.Sleep(time.Second)
			return 10000
		}, time.Second*2, arg, 1000, "hello world")

		rws = append(rws, resW)
	}

	fmt.Println("get result")
	for _, rw := range rws {
		go func(x *pool.ResultWorker) {
			fmt.Println("result: ", x.Get())
		}(rw)
	}
}
