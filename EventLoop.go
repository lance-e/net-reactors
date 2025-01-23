package netreactors

import (
	"fmt"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"sync/atomic"

	"github.com/lance-e/net-reactors/base/goroutine"
	"github.com/lance-e/net-reactors/base/util"
	"golang.org/x/sys/unix"
)

// ignaore the SIGPIPE automatically ,you don't need care here
var ignoreSIGPIPE = func() bool {
	signal.Ignore(syscall.SIGPIPE)
	return true
}()

const (
	kPollTimeMs = 10000
)

type EventLoop struct {
	looping_                int64 //atomic
	quit_                   int64 //atomic
	callingPendingFunctors_ int64 //atomic
	goroutineId_            int64
	wakeupFd_               int
	poller_                 Poller
	timerQueue_             *TimerQueue
	wakeupChannel           *Channel
	activeChannels          []*Channel
	mutex_                  sync.Mutex
	pendingFunctors_        []Functor //GuardedBy mutex_
}

// *************************
// public:
// *************************

func NewEventLoop() (el *EventLoop) {

	el = &EventLoop{
		looping_:                0,
		quit_:                   0,
		callingPendingFunctors_: 0,
		goroutineId_:            goroutine.GetGoid(),
		activeChannels:          make([]*Channel, 0),
		mutex_:                  sync.Mutex{},
		pendingFunctors_:        make([]Functor, 0),
	}
	//eventfd more sutable for wakeupFd_
	eventfd := util.CreateEventFd()
	if eventfd < 0 {
		Dlog.Printf("NewEventLoop: create eventfd failed , eventfd < 0 \n")
	}
	el.wakeupFd_ = eventfd

	//must set the  ownerLoop_ by transfer argument
	el.poller_ = NewDefaultPoller(el)
	el.timerQueue_ = NewTimerQueue(el)
	el.wakeupChannel = NewChannel(el, int32(eventfd))

	//wakeup callback
	el.wakeupChannel.SetReadCallback(el.HandleRead)
	el.wakeupChannel.EnableReading()
	return
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
		pollReturnTime := loop.poller_.Poll(kPollTimeMs, &loop.activeChannels)
		for i := 0; i < len(loop.activeChannels); i++ {
			loop.activeChannels[i].HandleEvent(pollReturnTime)
		}

		//the time to call DoPendingFunctors: make sure only in the event's callback that don't need to wakeup()
		loop.DoPendingFunctors()
	}

	fmt.Println("EventLoop stop looping")
	atomic.StoreInt64(&loop.looping_, 0)
}

func (loop *EventLoop) Quit() {
	atomic.StoreInt64(&loop.quit_, 1)
	if !loop.IsInLoopGoroutine() {
		loop.Wakeup()
	}
}

func (loop *EventLoop) AssertInLoopGoroutine() {
	if !loop.IsInLoopGoroutine() {
		loop.abortNotInLoopGoroutine()
	}
	return
}

func (loop *EventLoop) IsInLoopGoroutine() bool {
	return loop.goroutineId_ == goroutine.GetGoid()
}

func (loop *EventLoop) UpdateChannel(c *Channel) {
	if loop != c.loop_ {
		log.Panicf("loop.UpdateChannel: the chnnel's owner loop is this loop\n")
	}
	loop.AssertInLoopGoroutine()
	loop.poller_.UpdateChannel(c)
}

func (loop *EventLoop) RemoveChannel(c *Channel) {
	if loop != c.loop_ {
		log.Panicf("loop.RemoveChannel: the chnnel's owner loop is this loop\n")
	}
	loop.AssertInLoopGoroutine()
	//todo add eventHandling

	loop.poller_.RemoveChannel(c)
}

// make sure run in loop
func (loop *EventLoop) RunInLoop(cb Functor) {
	if loop.IsInLoopGoroutine() {
		cb()
	} else {
		loop.QueueInLoop(cb)
	}
}

// functor queue
func (loop *EventLoop) QueueInLoop(cb Functor) {
	loop.mutex_.Lock()
	loop.pendingFunctors_ = append(loop.pendingFunctors_, cb) //fix: performance
	loop.mutex_.Unlock()

	if !loop.IsInLoopGoroutine() || loop.callingPendingFunctors_ == 1 {
		loop.Wakeup()
	}
}

// execute callback in queue
func (loop *EventLoop) DoPendingFunctors() {
	functorTemp := make([]Functor, len(loop.pendingFunctors_))
	atomic.StoreInt64(&loop.callingPendingFunctors_, 1)

	loop.mutex_.Lock()
	//avoid to operate the critical section directly
	copy(functorTemp, loop.pendingFunctors_)
	loop.pendingFunctors_ = loop.pendingFunctors_[:0]
	loop.mutex_.Unlock()

	for i := 0; i < len(functorTemp); i++ {
		functorTemp[i]()
	}

	atomic.StoreInt64(&loop.callingPendingFunctors_, 0)
}

func (loop *EventLoop) Cancel(timerid TimerId) {
	loop.timerQueue_.Cancel(timerid)
}

func (loop *EventLoop) Wakeup() {
	var one uint64 = 1
	n, err := unix.Write(loop.wakeupFd_, []byte{
		byte(one), byte(one >> 8), byte(one >> 16), byte(one >> 24), byte(one >> 32), byte(one >> 40), byte(one >> 48), byte(one >> 56),
	})
	if n != 8 || err != nil {
		Dlog.Printf("Wakeup: write to wakeupFd_ failed, write %d bytes\n", n)
	}
}

// callback run at 't'
func (loop *EventLoop) RunAt(t time.Time, cb TimerCallback) TimerId {
	return loop.timerQueue_.AddTimer(cb, t, 0.0)
}

// callback run 'delay' from now
func (loop *EventLoop) RunAfter(duration time.Duration, cb TimerCallback) TimerId {
	return loop.timerQueue_.AddTimer(cb, time.Now().Add(duration), 0.0)
}

// callback run every 'interval'
func (loop *EventLoop) RunEvery(interval float64, cb TimerCallback) TimerId {
	return loop.timerQueue_.AddTimer(cb, time.Now().Add(time.Duration(interval)), interval)
}

// *************************
// private:
// *************************

func (loop *EventLoop) abortNotInLoopGoroutine() {
	log.Panicln("Abort Not In Loop Goroutine")
}

// wake up
func (loop *EventLoop) HandleRead(t time.Time) {
	var buf [8]byte
	//will read out the number of write event from eventfd's counter   into buf
	//such as: 5 times of write to eventfd , will read 5 ,and set the eventfd's counter to zere
	n, err := unix.Read(loop.wakeupFd_, buf[:])
	if err != nil || n != 8 {
		Dlog.Printf("loop.HandleRead: read from wakeupFd_ failed , read %d bytes\n", n)
	}
}
