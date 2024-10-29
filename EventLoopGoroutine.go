package netreactors

import (
	"sync"
)

type GoroutineCallback func(*EventLoop)

type EventLoopGoroutine struct {
	loop_     *EventLoop
	mutex_    *sync.Mutex
	cond_     *sync.Cond
	exiting_  bool
	callback_ GoroutineCallback
}

// *************************
// public:
// *************************

func NewEventLoopGoroutine(cb GoroutineCallback) (elg EventLoopGoroutine) {
	elg = EventLoopGoroutine{
		loop_:     nil,
		mutex_:    &sync.Mutex{},
		exiting_:  false,
		callback_: cb,
	}
	elg.cond_ = sync.NewCond(elg.mutex_)
	return elg
}

func (elg *EventLoopGoroutine) StartLoop() *EventLoop {
	//sub goroutine
	go elg.goroutineFunc()

	elg.mutex_.Lock()
	for elg.loop_ == nil {
		elg.cond_.Wait()
	}
	elg.mutex_.Unlock()

	return elg.loop_
}

// *************************
// private:
// *************************

func (elg *EventLoopGoroutine) goroutineFunc() {
	loop := NewEventLoop()

	if elg.callback_ != nil {
		elg.callback_(loop)
	}

	elg.mutex_.Lock()
	elg.loop_ = loop
	elg.cond_.Signal()
	elg.mutex_.Unlock()

	//event loop in sub goroutine
	elg.loop_.Loop()
}
