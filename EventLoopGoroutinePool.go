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

func NewEventLoopGoroutinePool(baseLoop *EventLoop, numGoroutine int, cb GoroutineCallback) (elgp *EventLoopGoroutinePool) {
	if numGoroutine < 0 {
		numGoroutine = 0
	}
	elgp = &EventLoopGoroutinePool{
		baseLoop_:     baseLoop,
		started_:      false,
		numGoroutine_: numGoroutine,
		next_:         0,
	}
	for i := 0; i < numGoroutine; i++ {
		t := NewEventLoopGoroutine(cb)
		elgp.goroutines_ = append(elgp.goroutines_, t)
		elgp.loops_ = append(elgp.loops_, t.loop_) //start loop and store the loop!
	}
	if numGoroutine == 0 && cb != nil {
		cb(baseLoop)
	}
	return
}

/* func (ep *EventLoopGoroutinePool) SetGoroutineNum(num int) { */
/* ep.numGoroutine_ = num */
/* } */

func (ep *EventLoopGoroutinePool) Start() {
	if ep.IsStarted() {
		log.Panicf("EventLoopGoroutinePool.Start(): failed , the pool is started\n")
	}
	// ep.baseLoop_.AssertInLoopGoroutine()

	ep.started_ = true

	for _, goroutineFunc := range ep.goroutines_ {
		goroutineFunc.StartLoop()
	}
}

func (ep *EventLoopGoroutinePool) GetNextLoop() (loop *EventLoop) {
	if !ep.IsStarted() {
		log.Panicf("EventLoopGoroutinePool.GetNextLoop(): failed , the pool isn't started\n")
	}
	// ep.baseLoop_.AssertInLoopGoroutine()

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
