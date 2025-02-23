package netreactors

import (
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

type (
	pollFdList []unix.PollFd        //for Poll
	eventList  []syscall.EpollEvent //for Epoll
	channelMap map[int32]*Channel
)

type Poller interface {
	Poll(timeoutMs int, activeChannels *[]*Channel) time.Time
	UpdateChannel(channel *Channel)
	RemoveChannel(channel *Channel)
	fillActiveChannels(numEvents int, activeChannels *[]*Channel)
}

func NewDefaultPoller(loop *EventLoop) Poller {
	return NewEpollPoller(loop)
}
