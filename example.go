package main

import (
	"fmt"
	"github.com/josexy/gopool/pool"
	"runtime"
	"sync/atomic"
)

var cnt int64

const SIZE = 10000000

func main() {
	p := pool.NewPool(runtime.NumCPU())
	defer func() { fmt.Println(cnt) }()
	defer p.Shutdown()

	for i := 0; i < runtime.NumCPU(); i++ {
		p.Submit(func() interface{} {
			for j := 0; j < SIZE; j++ {
				atomic.AddInt64(&cnt, 1)
			}
			return nil
		})
	}
}
