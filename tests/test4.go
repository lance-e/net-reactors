package main

import (
	"fmt"
	netreactors "net-reactors"
	"net-reactors/base/goroutine"
	"os"
	"time"
)

var gloop4 *netreactors.EventLoop
var count int

func printTid() {
	fmt.Printf("pid = %d , goid = %d \n ", os.Getpid(), goroutine.GetGoid())
	fmt.Printf("now : %s\n", time.Now().Format("2006-01-02 15:04:05"))
}

func bindPrint(msg string) func() {
	return func() {
		print(msg)
	}
}
func print(msg string) {
	fmt.Printf("msg : %s , %s \n", time.Now().Format("2006-01-02 15:04:05"), msg)
	count++
	if count >= 20 {
		gloop4.Quit()
	}
}

func main() {
	printTid()
	loop := netreactors.NewEventLoop()
	gloop4 = loop

	fmt.Printf("main:\n")

	loop.RunAfter(time.Second, bindPrint("once1"))

	loop.Loop()
	fmt.Printf("EventLoop stop looping\n")
	time.Sleep(1)
}
