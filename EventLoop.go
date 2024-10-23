package netreactors

import (
	"fmt"
	"log"
	"net-reactors/base/goroutine"
	"sync/atomic"
)

const (
	kPollTimeMs = 10000
)

type EventLoop struct {
	looping_       int64 //atomic
	quit_          int64 //atomic
	goroutineId_   int64
	poller_        *Poller
	activeChannels []*Channel
}

// *************************
// public:
// *************************

func NewEventLoop() EventLoop {
	id := goroutine.GetGoid()
	return EventLoop{
		looping_:     0,
		goroutineId_: id,
	}
}

// Loop
func (loop *EventLoop) Loop() {
	loop.AssertInLoopGoroutine()
	if !atomic.CompareAndSwapInt64(&loop.looping_, 0, 1) {
		panic("EventLoop is looping, should stop")
	}
	atomic.StoreInt64(&loop.quit_, 0)

	for loop.quit_ == 0 {
		loop.activeChannels = loop.activeChannels[:0]
		loop.poller_.Poll(kPollTimeMs, &loop.activeChannels)
		for i := 0; i < len(loop.activeChannels); i++ {
			loop.activeChannels[i].HandleEvent()
		}
	}

	fmt.Println("EventLoop stop looping")
	atomic.StoreInt64(&loop.looping_, 0)
}

func (loop *EventLoop) Quit() {
	atomic.StoreInt64(&loop.quit_, 1)
}

func (loop *EventLoop) AssertInLoopGoroutine() {
	if loop.goroutineId_ != goroutine.GetGoid() {
		loop.abortNotInLoopGoroutine()
	}
	fmt.Println("id is same")
	return
}

/* func (loop *EventLoop) IsInLoopGoroutine() bool { */
/* return loop.goroutineId_ == goroutine.GetGoid() */
/* } */

func (loop *EventLoop) UpdateChannel(c *Channel) {
	//fix: determine c.loop ?= loop
	loop.AssertInLoopGoroutine()
	loop.poller_.UpdateChannel(c)
}

// *************************
// private:
// *************************

func (loop *EventLoop) abortNotInLoopGoroutine() {
	log.Panicln("Abort Not In Loop Goroutine")
}
