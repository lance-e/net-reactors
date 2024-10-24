package main

import (
	"fmt"
	netreactors "net-reactors"

	"golang.org/x/sys/unix"
)

var gLoop3 *netreactors.EventLoop

func timeout() {
	fmt.Printf("Timeout!!!\n")
	gLoop3.Quit()
}

func main() {
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
