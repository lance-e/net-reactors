package netreactors

import (
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
	callback_   TimerCallback
	expiration_ time.Time
	interval_   float64
	repeat_     bool
	sequence_   int64
}

// *************************
// public:
// *************************

func NewTimer(cb TimerCallback, when time.Time, interval float64) *Timer {
	return &Timer{
		callback_:   cb,
		expiration_: when,
		interval_:   interval,
		repeat_:     interval > 0.0,
		sequence_:   atomic.AddInt64(&AtomicNumber, 1),
	}
}

func (t *Timer) Run() {
	t.callback_()
}

func (t *Timer) Expiration() time.Time {
	return t.expiration_
}

func (t *Timer) Repeat() bool {
	return t.repeat_
}
func (t *Timer) Sequence() int64 {
	return t.sequence_
}

func (t *Timer) Restart(now time.Time) {
	if t.repeat_ {
		t.expiration_ = now.Add(time.Duration(t.interval_))
	} else {
		t.expiration_ = time.Time{}
	}
}

// *************************
// private:
// *************************
