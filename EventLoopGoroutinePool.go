package netreactors

import (
	"log"
)

type EventLoopGoroutinePool struct {
	baseLoop_     *EventLoop
	name_         string
	started_      bool
	numGoroutine_ int
	next_         int
	goroutines_   []*EventLoopGoroutine
	loops_        []*EventLoop
}

// *************************
// public:
// *************************

func NewEventLoopGoroutinePool(baseLoop *EventLoop) *EventLoopGoroutinePool {
	return &EventLoopGoroutinePool{
		baseLoop_:     baseLoop,
		started_:      false,
		numGoroutine_: 1,
		next_:         0,
		goroutines_:   make([]*EventLoopGoroutine, 0),
		loops_:        make([]*EventLoop, 0),
	}
}

func (ep *EventLoopGoroutinePool) SetGoroutineNum(num int) {
	ep.numGoroutine_ = num
}

func (ep *EventLoopGoroutinePool) Start(cb GoroutineCallback) {
	if ep.IsStarted() {
		log.Panicf("EventLoopGoroutinePool.Start(): failed , the pool is started\n")
	}
	ep.baseLoop_.AssertInLoopGoroutine()

	ep.started_ = true
	for i := 0; i < ep.numGoroutine_; i++ {
		t := NewEventLoopGoroutine(cb)
		ep.goroutines_ = append(ep.goroutines_, t)
		ep.loops_ = append(ep.loops_, t.StartLoop()) //start loop and store the loop!
	}

	if ep.numGoroutine_ == 0 && cb != nil {
		cb(ep.baseLoop_)
	}

}

func (ep *EventLoopGoroutinePool) GetNextLoop() (loop *EventLoop) {
	if !ep.IsStarted() {
		log.Panicf("EventLoopGoroutinePool.GetNextLoop(): failed , the pool isn't started\n")
	}
	ep.baseLoop_.AssertInLoopGoroutine()

	loop = ep.baseLoop_
	if len(ep.loops_) > 0 {
		loop = ep.loops_[ep.next_]
		ep.next_++
		if ep.next_ >= len(ep.loops_) {
			ep.next_ = 0
		}
	}
	return
}

func (ep *EventLoopGoroutinePool) IsStarted() bool {
	return ep.started_
}

// *************************
// private:
// *************************
