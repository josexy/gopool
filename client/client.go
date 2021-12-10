package main

import (
	"fmt"
	"io/ioutil"
	"net"
)

func main() {
	for i := 0; i < 1000; i++ {
		resp, err := net.Dial("tcp", ":6666")
		if err != nil {
			fmt.Println(err)
			continue
		}
		resp.Write([]byte("hello server"))
		bs, err := ioutil.ReadAll(resp)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(string(bs))
		}
		resp.Close()
	}
}
