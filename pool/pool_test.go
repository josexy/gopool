package pool

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

const M = 10000000

var counter int64

func GetHtml(url string, t *testing.T) interface{} {
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	t.Log(resp.StatusCode, len(data))
	return data
}

func TestHttpGetWithPool(t *testing.T) {
	url := "https://www.example.com"
	p := NewPool(runtime.NumCPU())
	for i := 0; i < 20; i++ {
		p.Submit(func(...interface{}) interface{} {
			_ = GetHtml(url, t)
			return nil
		}, nil)
	}
	p.Shutdown()
}

func TestSum(t *testing.T) {
	p := NewPool(runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		p.Submit(func(...interface{}) interface{} {
			for j := 0; j < M; j++ {
				atomic.AddInt64(&counter, 1)
			}
			return nil
		}, nil)
	}
	p.Shutdown()
	fmt.Println(counter)
}

func TestWorker_Get(t *testing.T) {
	p := NewPool(2)
	w1 := p.Submit(func(...interface{}) interface{} {
		time.Sleep(time.Second * 5)
		t.Log("Work1 done")
		return nil
	}, nil)

	w2 := p.SubmitResult(func(...interface{}) interface{} {
		time.Sleep(time.Second * 3)
		t.Logf("Work2 done")
		return 10000
	}, time.Second*5, nil)

	w3 := p.SubmitResult(func(...interface{}) interface{} {
		time.Sleep(time.Second * 3)
		t.Log("Work3 done")
		return 20000
	}, 0, nil)

	t.Log("work1:", w1.Get())
	t.Log("work2:", w2.Get())
	t.Log("work3:", w3.Get())
	p.Shutdown()
}

func TestPool_Shutdown(t *testing.T) {
	p := NewPool(8)
	defer p.Shutdown()

	for i := 0; i < 10; i++ {
		p.Submit(func(...interface{}) interface{} {
			time.Sleep(time.Second * 3)
			return nil
		}, nil)
	}
}

func TestResultWorker(t *testing.T) {
	p := NewPool(runtime.NumCPU())
	defer p.Shutdown()

	var wgr []IWorker
	for i := 0; i < runtime.NumCPU(); i++ {
		worker := p.SubmitResult(func(...interface{}) interface{} {
			var n int
			for j := 0; j < 900000000; j++ {
				n++
			}
			return n
		}, 0, nil)
		wgr = append(wgr, worker)
	}

	for _, worker := range wgr {
		t.Log(worker.Get())
	}
}
