package netreactors

import (
	"container/heap"
	"log"
	"sort"
	"sync/atomic"
	"time"

	"github.com/lance-e/net-reactors/base/util"
)

// fix: more high performance
type TimerQueue struct {
	loop_                 *EventLoop
	timeChannel_          *Channel
	timerfd_              int
	timers_               TimerList //timer set
	activeTimers_         map[*Timer]int64
	cancelingTimers_      map[*Timer]int64
	callingExpiredTimers_ int64 //atomic
}

//*************************
//public:
//*************************

func NewTimerQueue(loop *EventLoop) (tq *TimerQueue) {
	tfd := util.CreateTimerFd()
	if tfd < 0 {
		log.Printf("NewTimerQueue: create timerfd failed, timerfd < 0 \n")
	}

	tq = &TimerQueue{
		loop_:                 loop,
		timeChannel_:          NewChannel(loop, int32(tfd)),
		timerfd_:              tfd,
		timers_:               make(TimerList, 0),
		activeTimers_:         make(map[*Timer]int64, 0),
		cancelingTimers_:      make(map[*Timer]int64, 0),
		callingExpiredTimers_: 0,
	}

	//set timer callback
	tq.timeChannel_.SetReadCallback(tq.HandleRead)
	tq.timeChannel_.EnableReading()
	return
}

// called when timer alarms
func (tq *TimerQueue) HandleRead(t time.Time) {
	tq.loop_.AssertInLoopGoroutine()
	now := time.Now()
	util.ReadTimerfd(tq.timerfd_, now)

	//handle expirated timer
	expired := tq.getExpired(now)
	if len(expired) == 0 {
		panic("getExpired failed, get nil expired")
	}

	atomic.StoreInt64(&tq.callingExpiredTimers_, 1)
	//clear cancelingTimers_
	tq.cancelingTimers_ = make(map[*Timer]int64, 0)

	for _, entry := range expired {
		entry.timer_.Run()
	}
	atomic.StoreInt64(&tq.callingExpiredTimers_, 0)

	//reset expired timers
	tq.reset(expired, now)
}
func (tq *TimerQueue) AddTimer(cb TimerCallback, when time.Time, interval float64) TimerId {
	timer := NewTimer(cb, when, interval)
	tq.loop_.RunInLoop(tq.bindAddTimerInLoop(timer))
	return NewTimerId(timer, timer.sequence_)
}

func (tq *TimerQueue) Cancel(timerid TimerId) {
	tq.loop_.RunInLoop(tq.bindCancelTimerInLoop(timerid))
}

//*************************
//private:
//*************************

// move out all expired timers from set
func (tq *TimerQueue) getExpired(now time.Time) []TimeEntry {
	if tq.timers_.Len() != len(tq.activeTimers_) {
		log.Panicf("getExpired: the length of timers_ and activeTimers_ is not same\n")
	}
	expired := make([]TimeEntry, 0)
	//binary search expired timer
	idx := sort.Search(len(tq.timers_), func(i int) bool {
		return !tq.timers_[i].when_.Before(now) //equal to now <= tq.timers_[i].when
	})

	//copy the expired timer
	expired = append(expired, tq.timers_[:idx]...)
	//remove the expired timer
	tq.timers_ = tq.timers_[idx:]

	//remove expired timer from activeTimers_
	for _, entry := range expired {
		delete(tq.activeTimers_, entry.timer_)
	}
	if tq.timers_.Len() != len(tq.activeTimers_) {
		log.Panicf("getExpired: the length of timers_ and activeTimers_ is not same\n")
	}
	return expired
}

// bind the timer that don't need to transfer the argument
func (tq *TimerQueue) bindAddTimerInLoop(timer *Timer) func() {
	return func() {
		tq.addTimerInLoop(timer)
	}
}

// goroutine safe
func (tq *TimerQueue) addTimerInLoop(timer *Timer) {
	tq.loop_.AssertInLoopGoroutine()
	earliestChaned := tq.insert(timer)
	if earliestChaned {
		util.ResetTimerfd(tq.timerfd_, timer.expiration_)
	}
}

// bind the timerid
func (tq *TimerQueue) bindCancelTimerInLoop(timerid TimerId) func() {
	return func() {
		tq.cancelTimerInLoop(timerid)
	}
}

// goroutine safe
func (tq *TimerQueue) cancelTimerInLoop(timerid TimerId) {
	tq.loop_.AssertInLoopGoroutine()
	if tq.timers_.Len() != len(tq.activeTimers_) {
		log.Panicf("TimerQueue.cancelTimerInLoop(): the length of timers_ and activeTimers_ is not same\n")
	}

	if _, ok := tq.activeTimers_[timerid.Timer_]; ok {
		for i, v := range tq.timers_ {
			if v.timer_.Expiration() == timerid.Timer_.Expiration() && v.timer_.Sequence() == timerid.Sequence_ {
				heap.Remove(&tq.timers_, i)
			}
		}
		delete(tq.activeTimers_, timerid.Timer_)
	} else if atomic.LoadInt64(&tq.callingExpiredTimers_) == 1 {
		tq.cancelingTimers_[timerid.Timer_] = timerid.Sequence_
	}

	if tq.timers_.Len() != int(len(tq.activeTimers_)) {
		log.Panicf("TimerQueue.cancelTimerInLoop(): the length of timers_ and activeTimers_ is not same\n")
	}
}

func (tq *TimerQueue) reset(expired []TimeEntry, now time.Time) {
	var nextExpired time.Time
	//handle expired timer:
	//if it was repeatly and not in cancelingTimers_ that will restart
	//otherwise it will be delete
	for _, entry := range expired {
		if _, ok := tq.cancelingTimers_[entry.timer_]; !ok && entry.timer_.Repeat() {
			entry.timer_.Restart(now)
			tq.insert(entry.timer_)
		} else {
			delete(tq.cancelingTimers_, entry.timer_)
			entry.timer_ = nil
		}
	}
	if tq.timers_.Len() > 0 {
		nextExpired = tq.timers_[0].timer_.expiration_
	}
	if !nextExpired.IsZero() {
		util.ResetTimerfd(tq.timerfd_, nextExpired)
	}
}

func (tq *TimerQueue) insert(timer *Timer) bool {
	tq.loop_.AssertInLoopGoroutine()
	if tq.timers_.Len() != len(tq.activeTimers_) {
		log.Panicf("insert: the length of timers_ and activeTimers_ is not same\n")
	}

	earliestChanged := false
	when := timer.Expiration()
	if tq.timers_.Len() == 0 || when.Before(tq.timers_[0].when_) {
		earliestChanged = true
	}
	entry := TimeEntry{
		when_:  when,
		timer_: timer,
	}
	//insert
	heap.Push(&tq.timers_, entry)
	tq.activeTimers_[timer] = timer.Sequence()
	if tq.timers_.Len() != len(tq.activeTimers_) {
		log.Panicf("insert: the length of timers_ and activeTimers_ is not same\n")
	}

	return earliestChanged
}
