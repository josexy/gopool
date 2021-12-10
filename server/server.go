package main

import (
	"bufio"
	"fmt"
	"net"
	"runtime"

	"github.com/josexy/gopool/pool"
)

func handleConn(con net.Conn) {
	fmt.Println(con.RemoteAddr())
	rw := bufio.NewReadWriter(bufio.NewReader(con), bufio.NewWriter(con))
	buf := make([]byte, 1024)
	n, err := rw.Read(buf[:])
	if err != nil {
		fmt.Println(err)
		con.Close()
		return
	}
	fmt.Println(string(buf[:n]))
	rw.Write([]byte("hello client"))
	rw.Flush()
	con.Close()
}

func main() {
	p := pool.NewPool(runtime.NumCPU())
	defer p.Shutdown()

	listener, err := net.Listen("tcp", ":6666")
	defer func() { _ = listener.Close() }()

	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		p.Submit(func(arg ...interface{}) interface{} {
			handleConn(arg[0].(net.Conn))
			return nil
		}, conn)
	}
}
