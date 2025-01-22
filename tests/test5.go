package main

import (
	"fmt"
	"os"
	"time"

	netreactors "github.com/lance-e/net-reactors"
)

var globalLoop5 *netreactors.EventLoop
var g_flag = 0

func run4() {
	fmt.Printf("run4: pid:%d , flag:%d\n", os.Getpid(), g_flag)
	globalLoop5.Quit()
}
func run3() {
	fmt.Printf("run3: pid:%d , flag:%d\n", os.Getpid(), g_flag)
	globalLoop5.RunAfter(3*time.Second, run4)
	g_flag = 3
}

func run2() {
	fmt.Printf("run2: pid:%d , flag:%d\n", os.Getpid(), g_flag)
	globalLoop5.QueueInLoop(run3)
}
func run1() {
	g_flag = 1
	fmt.Printf("run1: pid:%d , flag:%d\n", os.Getpid(), g_flag)
	globalLoop5.RunInLoop(run2)
	g_flag = 2
}

func main5() {
	loop := netreactors.NewEventLoop()
	globalLoop5 = loop

	loop.RunAfter(2, run1)
	loop.Loop()

	fmt.Printf("main: pid:%d , flag:%d\n", os.Getpid(), g_flag)

}
