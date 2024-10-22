package net

import (
	"fmt"
	"net-reactors/base/goroutine"

	"golang.org/x/sys/unix"
)

type EventLoop struct {
	looping_    bool
	goroutineId int64
}

func NewEventLoop() EventLoop {
	return EventLoop{
		looping_:    false,
		goroutineId: goroutine.GetGoid(),
	}
}

// Loop
func (loop *EventLoop) Loop() {
	loop.AssertInLoopThread()
	loop.looping_ = true

	fds := []unix.PollFd{}
	_, _ = unix.Poll(fds, 5*1000)

	fmt.Println("EventLoop stop looping")
	loop.looping_ = false
}

func (loop *EventLoop) AssertInLoopThread() {
	if !loop.IsInLoopThread() {
		loop.abortNotInLoopThread()
	}
}

func (loop *EventLoop) IsInLoopThread() bool {
	return loop.goroutineId == goroutine.GetGoid()
}

func (loop *EventLoop) abortNotInLoopThread() {
	fmt.Println("Abort Not In Loop Thread")
	panic("Abort")
}
