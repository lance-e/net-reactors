package main

import (
	"sync"

	netreactors "github.com/lance-e/net-reactors"
)

var globalLoop2 *netreactors.EventLoop
var wg sync.WaitGroup

func goroutineFunc2() {
	globalLoop2.Loop()
	wg.Done()
}

func main2() {
	loop := netreactors.NewEventLoop()
	globalLoop2 = loop

	wg.Add(1)
	go goroutineFunc2()

	wg.Wait()
}
