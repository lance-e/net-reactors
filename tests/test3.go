package main

import (
	"fmt"
	"time"

	netreactors "github.com/lance-e/net-reactors"
	"golang.org/x/sys/unix"
)

var gLoop3 *netreactors.EventLoop

func timeout(t time.Time) {
	fmt.Printf("%s timeout!\n", t.Format("2006-01-02 15:04:05"))
	gLoop3.Quit()
}

func main3() {
	fmt.Printf("%s started!\n", time.Now().Format("2006-01-02 15:04:05"))
	loop := netreactors.NewEventLoop()
	gLoop3 = loop
	timerfd, err := unix.TimerfdCreate(unix.CLOCK_MONOTONIC, unix.TFD_NONBLOCK|unix.TFD_CLOEXEC)
	if err != nil {
		panic("create timerfd failed")
	}

	fmt.Printf("In main: EventLoop address is %d\n", &loop)

	ch := netreactors.NewChannel(loop, int32(timerfd))
	ch.SetReadCallback(timeout)
	ch.EnableReading()

	howlong := unix.ItimerSpec{}
	howlong.Value.Sec = 5
	unix.TimerfdSettime(timerfd, 0, &howlong, nil)

	loop.Loop()

	unix.Close(timerfd)
}
