package netreactors

import (
	"fmt"
	"net-reactors/base/util"

	"golang.org/x/sys/unix"
)

const (
	kNoneEvent  = 0
	kReadEvent  = unix.POLLIN | unix.POLLPRI
	kWriteEvent = unix.POLLOUT
)

type Channel struct {
	loop_ *EventLoop

	fd_      int32
	events_  int16
	revents_ int16
	index_   int

	readCallback_  util.EventCallback
	writeCallback_ util.EventCallback
	errorCallback_ util.EventCallback
}

//*************************
//public:
//*************************

func NewChannel(loop *EventLoop, fdArg int32) Channel {
	return Channel{
		loop_:    loop,
		fd_:      fdArg,
		events_:  0,
		revents_: 0,
		index_:   -1,
	}
}

func (c *Channel) HandleEvent() {
	//determine whether fd is valid
	if c.revents_&unix.POLLNVAL != 0 {
		fmt.Printf("WARN: Channel.HandleEvent POLLNVAL\n")
	}
	if c.revents_&(unix.POLLERR|unix.POLLNVAL) != 0 {
		if c.errorCallback_ != nil {
			c.errorCallback_()
		}
	}
	if c.revents_&(unix.POLLIN|unix.POLLPRI|unix.POLLRDHUP) != 0 {
		if c.readCallback_ != nil {
			c.readCallback_()
		}
	}
	if c.revents_&(unix.POLLOUT) != 0 {
		if c.writeCallback_ != nil {
			c.writeCallback_()
		}
	}
}

func (c *Channel) SetReadCallback(cb util.EventCallback) { //fix
	c.readCallback_ = cb
}

func (c *Channel) SetWriteCallback(cb util.EventCallback) { //fix
	c.writeCallback_ = cb
}

func (c *Channel) SetErrorCallback(cb util.EventCallback) { //fix
	c.errorCallback_ = cb
}
func (c *Channel) Fd() int32 {
	return c.fd_
}

func (c *Channel) events() int16 {
	return c.events_
}

func (c *Channel) SetRevents(revt int16) {
	c.revents_ = revt
}

func (c *Channel) IsNoneEvent() bool {
	return c.events_ == kNoneEvent
}
func (c *Channel) EnableReading() {
	c.events_ |= kReadEvent
	c.update()
}

/* func (c *Channel) EnableWriting() { */
/* c.events_ |= kWriteEvent */
/* c.update() */
/* } */
/* func (c *Channel) DisableWriting() { */
/* c.events_ &= ^kWriteEvent */
/* c.update() */
/* } */
/* func (c *Channel) DisableAll() { */
/* c.events_ = kNoneEvent */
/* c.update() */
/* } */
func (c *Channel) Index() int {
	return c.index_
}
func (c *Channel) SetIndex(idx int) {
	c.index_ = idx
}

func (c *Channel) OwnerLoop() *EventLoop {
	return c.loop_
}

// *************************
// private:
// *************************
func (c *Channel) update() {
	c.loop_.UpdateChannel(c)
}