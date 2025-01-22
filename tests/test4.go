package main

import (
	"fmt"
	"os"
	"time"

	netreactors "github.com/lance-e/net-reactors"
	"github.com/lance-e/net-reactors/base/goroutine"
)

var gloop4 *netreactors.EventLoop
var count int
var toCancel netreactors.TimerId

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

func cancleSelf() {
	fmt.Printf("cancleSelf\n")
	gloop4.Cancel(toCancel)
}

func main4() {
	printTid()
	loop := netreactors.NewEventLoop()
	gloop4 = loop

	fmt.Printf("main:\n")

	loop.RunAfter(time.Duration(10*(time.Second)), bindPrint("once10"))
	loop.RunAfter(time.Second, bindPrint("once1"))
	loop.RunAfter(time.Duration(1.5*float64(time.Second)), bindPrint("once1.5"))
	loop.RunAfter(time.Duration(2.5*float64(time.Second)), bindPrint("once2.5"))
	loop.RunAfter(time.Duration(3.5*float64(time.Second)), bindPrint("once3.5"))
	loop.RunEvery(float64(2*time.Second), bindPrint("every2"))
	loop.RunEvery(float64(3*time.Second), bindPrint("every3"))

	// toCancel = loop.RunEvery(float64(5*time.Second), cancleSelf)

	loop.Loop()
	fmt.Printf("main EventLoop stop looping\n")
	time.Sleep(1)
}
