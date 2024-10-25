package netreactors

import (
	"log"
	"time"

	"golang.org/x/sys/unix"
)

type TimeEntry struct {
	time_  time.Time
	timer_ *Timer
}

// fix: more high performance
type TimerQueue struct {
	loop_                 *EventLoop
	timeChannel_          *Channel
	timerfd_              int
	timers_               map[*TimeEntry]struct{} //timer set
	callingExpiredTimers_ int64                   //atomic
	//activeTimer
	//cancelingTimer
}

//*************************
//public:
//*************************

func NewTimerQueue(loop *EventLoop) (tq *TimerQueue) {
	//timerfd
	tfd, err := unix.TimerfdCreate(unix.CLOCK_MONOTONIC, unix.TFD_NONBLOCK|unix.TFD_CLOEXEC)
	if err != nil {
		log.Panicf("NewTimerQueue: TimerfdCreate failed\n")
	}

	tq = &TimerQueue{
		loop_:                 loop,
		timeChannel_:          NewChannel(loop, int32(tfd)),
		timerfd_:              tfd,
		timers_:               map[*TimeEntry]struct{}{},
		callingExpiredTimers_: 0,
	}

	//set timer callback
	tq.timeChannel_.SetReadCallback(tq.HandleRead)
	tq.timeChannel_.EnableReading()
	return
}

// called when timer alarms
func (tq *TimerQueue) HandleRead() {
	tq.loop_.AssertInLoopGoroutine()
	now := time.Now()

	//read timefd
	var buf [8]byte
	n, err := unix.Read(tq.timerfd_, buf[:])
	howmany := uint64(buf[0]) | uint64(buf[1])<<8 | uint64(buf[2])<<16 | uint64(buf[3])<<24 | uint64(buf[4])<<32 | uint64(buf[5])<<40 | uint64(buf[6])<<48 | uint64(buf[7])<<56
	log.Printf("tq.HandleRead expirated %d times at now:%s\n ", howmany, now.Format("2006-01-02 15:04:05"))
	if err != nil || n != 8 {
		log.Printf("tq.HandleRead: read timerfd failed, read %d bytes instead of 8 bytes\n", n)
	}

	//handle expirated timer

}

//*************************
//private:
//*************************

func (tq *TimerQueue) getExpired(now time.Time) {
	//expired := make([]*TimeEntry, 0)

}
func (tq *TimerQueue) reset(expired []*TimeEntry, now time.Time) {

}
func (tq *TimerQueue) insert(timer *Timer) {

}
