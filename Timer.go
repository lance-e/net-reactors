package netreactors

import (
	"net-reactors/base/util"
	"sync/atomic"
	"time"
)

var AtomicNumber int64 = 0

//TimerId

type TimerId struct {
	Timer_    *Timer
	Sequence_ int64
}

func NewTimerId(timer *Timer, seq int64) TimerId {
	return TimerId{
		Timer_:    timer,
		Sequence_: seq,
	}
}

//Timer

type Timer struct {
	callback_   util.TimerCallback
	expiration_ time.Time
	interval_   float64
	repeat_     bool
	sequence_   int64
}

// *************************
// public:
// *************************

func NewTimer(cb util.TimerCallback, when time.Time, interval float64) Timer {
	return Timer{
		callback_:   cb,
		expiration_: when,
		interval_:   interval,
		repeat_:     false,
		sequence_:   atomic.AddInt64(&AtomicNumber, 1),
	}
}

func (t *Timer) Run() {
	t.callback_()
}

func (t *Timer) Expiration() time.Time {
	return t.expiration_
}

// *************************
// private:
// *************************
