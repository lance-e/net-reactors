package main

import (
	"net-reactors/net"
	"sync"
)

var globalLoop *net.EventLoop
var wg sync.WaitGroup

func goroutineFunc2() {
	globalLoop.Loop()
	wg.Done()
}

func main2() {
	loop := net.NewEventLoop()
	globalLoop = &loop

	wg.Add(1)
	go goroutineFunc2()

	wg.Wait()
}
