package main

import (
	"fmt"
	"net-reactors/base/goroutine"
	"net-reactors/net"
	"os"
)

func goroutineFunc() {
	fmt.Printf("goroutineFunc: pid:%d , gid:%d\n", os.Getpid(), goroutine.GetGoid())

	loop := net.NewEventLoop()
	loop.Loop()
}

func main1() {
	fmt.Printf("main: pid:%d , gid:%d\n", os.Getpid(), goroutine.GetGoid())

	mainLoop := net.NewEventLoop()
	go goroutineFunc()

	mainLoop.Loop()

	os.Exit(0)
}
