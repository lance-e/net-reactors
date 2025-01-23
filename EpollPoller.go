package netreactors

import (
	"log"
	"time"

	"golang.org/x/sys/unix"
)

const (
	InitEventListSize = 16
	kNew              = -1
	kAdded            = 1
	kDeleted          = 2
)

type EpollPoller struct {
	loop_     *EventLoop //event loop owner
	epollfd   int        //epollfd
	events_   eventList  //cathe of epoll event
	channels_ channelMap //map of fd and Channel
}

// *************************
// public:
// *************************

func NewEpollPoller(loop *EventLoop) (ep *EpollPoller) {
	ep = &EpollPoller{
		loop_:     loop,
		events_:   make(eventList, InitEventListSize),
		channels_: make(channelMap),
	}
	var err error
	if ep.epollfd, err = unix.EpollCreate1(unix.EPOLL_CLOEXEC); err != nil {
		log.Panicf("EpollPoller.NewEpollPoller(): create epollfd failed\n")
	}

	return
}

func (e *EpollPoller) Poll(timeoutMs int, activeChannels *[]*Channel) time.Time {
	n, err := unix.EpollWait(e.epollfd, e.events_, timeoutMs)
	if err != nil || n < 0 {
		if err != unix.EINTR { //ignore EINTR
			log.Panicf("EpollPoller.Poll failed ,n:%d , err:%s \n", n, err.Error())
		}
	}
	now := time.Now()
	if n > 0 {
		Dlog.Printf("%d events happended\n", n)
		e.fillActiveChannels(n, activeChannels)
		if n == len(e.events_) {
			e.events_ = make(eventList, len(e.events_), 2*cap(e.events_))
		}
	} else if n == 0 {
		Dlog.Printf("nothing happended\n")
	}
	return now
}

func (e *EpollPoller) UpdateChannel(channel *Channel) {
	e.AssertInLoopGoroutine()
	Dlog.Printf("UpdateChannel: fd=%d , events=%d\n", channel.fd_, channel.events_)
	index := channel.Index()
	if index == kNew || index == kDeleted {
		// a new one ,add with EPOLL_CTL_ADD
		if index == kNew {
			if _, ok := e.channels_[channel.Fd()]; ok {
				log.Panicln("EpollPoller.UpdateChannel: channel's fd already exist")
			}
			e.channels_[channel.Fd()] = channel
		} else {
			if _, ok := e.channels_[channel.Fd()]; !ok {
				log.Panicln("EpollPoller.UpdateChannel: channel's fd doesn't exist")
			}
			if e.channels_[channel.Fd()] != channel {
				log.Panicln("EpollPoller.UpdateChannel: the channel corresponding to this fd isn't some one")
			}
		}

		channel.SetIndex(kAdded)
		e.update(unix.EPOLL_CTL_ADD, channel)
	} else {
		// update existing one with EPOLL_CTL_MOD/DEL
		if _, ok := e.channels_[channel.Fd()]; !ok {
			log.Panicln("EpollPoller.UpdateChannel: channel's fd doesn't exist")
		}
		if e.channels_[channel.Fd()] != channel {
			log.Panicln("EpollPoller.UpdateChannel: the channel corresponding to this fd isn't some one")
		}
		if channel.Index() != kAdded {
			log.Panicln("EpollPoller.UpdateChannel: channel's index isn't kAdded")
		}
		if channel.IsNoneEvent() {
			channel.SetIndex(kDeleted)
			e.update(unix.EPOLL_CTL_DEL, channel)
		} else {
			e.update(unix.EPOLL_CTL_MOD, channel)
		}
	}
}

func (e *EpollPoller) RemoveChannel(channel *Channel) {
	e.AssertInLoopGoroutine()
	Dlog.Printf("RemoveChannel: fd = %d\n", channel.Fd())
	//assert
	if _, ok := e.channels_[channel.Fd()]; !ok {
		log.Panicf("EpollPoller.RemoveChannel:channel not found\n")
	}
	if e.channels_[channel.Fd()] != channel {
		log.Panicf("EpollPoller.RemoveChannel:channel isn't the target channel\n ")
	}
	if !channel.IsNoneEvent() {
		log.Panicf("EpollPoller.RemoveChannel:channel isn't none event\n")
	}
	idx := channel.Index()
	if idx != kAdded && idx != kDeleted {
		log.Panicln("EpollPoller.UpdateChannel: the index of channel is wrong")
	}

	delete(e.channels_, channel.Fd())
	if idx == kAdded {
		e.update(unix.EPOLL_CTL_DEL, channel)
	}
	channel.SetIndex(kNew)
}

func (e *EpollPoller) AssertInLoopGoroutine() {
	e.loop_.AssertInLoopGoroutine()
}

// *************************
// private:
// *************************

func (e *EpollPoller) fillActiveChannels(numEvents int, activeChannels *[]*Channel) {
	if numEvents > len(e.events_) {
		log.Panicf("EpollPoller.fillActiveChannels: numEvents more than events_\n")
	}
	for i := 0; i < numEvents; i++ {
		v, ok := e.channels_[e.events_[i].Fd]
		if !ok {
			log.Panicf("EpollPoller.fillActiveChannels: not found target in channelMap\n")
		}
		v.SetRevents(int16(e.events_[i].Events))
		*activeChannels = append(*activeChannels, v)
	}
}

func (e *EpollPoller) update(op int, channel *Channel) {
	event := unix.EpollEvent{
		Events: uint32(channel.events_),
		Fd:     channel.Fd(),
	}
	Dlog.Printf("EpollCtl: operation is [%s] , fd is [%d] \n", e.operationToString(op), channel.Fd())
	if err := unix.EpollCtl(e.epollfd, op, int(channel.Fd()), &event); err != nil {
		if op == unix.EPOLL_CTL_DEL {
			Dlog.Printf("EpollPoller.update() error: operation is [%s] , fd is [%d]\n", e.operationToString(op), channel.Fd())
		} else {
			log.Panicf("EpollPoller.update() panic: operation is [%s] , fd is [%d]\n", e.operationToString(op), channel.Fd())
		}
	}
}

func (e *EpollPoller) operationToString(operation int) string {
	switch operation {
	case unix.EPOLL_CTL_ADD:
		return "ADD"
	case unix.EPOLL_CTL_DEL:
		return "DEL"
	case unix.EPOLL_CTL_MOD:
		return "MOD"
	default:
		return "UNKNOWN"
	}
}
