package main

import (
	"fmt"
	netreactors "net-reactors"
	"net-reactors/base/goroutine"
	"os"
)

func goroutineFunc() {
	fmt.Printf("goroutineFunc: pid:%d , gid:%d\n", os.Getpid(), goroutine.GetGoid())

	loop := netreactors.NewEventLoop()
	loop.Loop()
}

func main1() {
	fmt.Printf("main: pid:%d , gid:%d\n", os.Getpid(), goroutine.GetGoid())

	mainLoop := netreactors.NewEventLoop()
	go goroutineFunc()

	mainLoop.Loop()

	os.Exit(0)
}
