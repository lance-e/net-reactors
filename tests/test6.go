package main

import (
	"fmt"
	"os"
	"time"

	netreactors "github.com/lance-e/net-reactors"
	"github.com/lance-e/net-reactors/base/goroutine"
)

func GoroutineFunc() {
	fmt.Printf("Func: pid:%d , goid:%d\n", os.Getpid(), goroutine.GetGoid())
}

func main6() {
	fmt.Printf("main: pid:%d , goid:%d\n", os.Getpid(), goroutine.GetGoid())

	eg := netreactors.NewEventLoopGoroutine(nil)
	loop := eg.StartLoop()

	loop.RunInLoop(GoroutineFunc)
	time.Sleep(time.Second)

	loop.RunAfter(3*time.Second, GoroutineFunc)
	time.Sleep(10 * time.Second)
	loop.Quit()

	fmt.Printf("main routine exit...\n")
}
