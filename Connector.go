package netreactors

import (
	"log"
	"net/netip"
	"sync/atomic"
	"time"

	"github.com/lance-e/net-reactors/base/socket"
	"golang.org/x/sys/unix"
)

const (
	MaxRetryTime  = 30 * time.Second
	InitRetryTime = 500 * time.Millisecond
)

type Connector struct {
	loop_                           *EventLoop
	serverAddr_                     *netip.AddrPort
	channel_                        *Channel
	connected_                      int64 //atomic
	state_                          int
	retryTime_                      time.Duration
	connectorNewConnectionCallback_ ConnectorNewConnectionCallback
}

// *************************
// public:
// *************************

func NewConnector(loop *EventLoop, addr *netip.AddrPort) *Connector {
	return &Connector{
		loop_:       loop,
		serverAddr_: addr,
		channel_:    nil,
		connected_:  0,
		state_:      kDisconnected,
		retryTime_:  InitRetryTime,
	}
}

func (c *Connector) SetConnectorNewConnectionCallback(cb ConnectorNewConnectionCallback) {
	c.connectorNewConnectionCallback_ = cb
}

func (c *Connector) Start() {
	atomic.StoreInt64(&c.connected_, 1)
	c.loop_.RunInLoop(c.startInLoop)
}

func (c *Connector) Restart() {
	c.loop_.AssertInLoopGoroutine()
	c.setState(kDisconnected)
	c.retryTime_ = InitRetryTime
	atomic.StoreInt64(&c.connected_, 1)
	c.startInLoop()
}

func (c *Connector) Stop() {
	atomic.StoreInt64(&c.connected_, 0)
	c.loop_.RunInLoop(c.stopInLoop)
}

// *************************
// private:
// *************************

func (c *Connector) startInLoop() {
	c.loop_.AssertInLoopGoroutine()
	if c.state_ != kDisconnected {
		log.Panicf("Connector.startInLoop() : failed because connector's state isn't kConnecting\n")
	}
	if atomic.LoadInt64(&c.connected_) == 1 {
		c.connect()
	} else {
		Dlog.Printf("Connector.startInLoop(): don't connect\n ")
	}

}

func (c *Connector) stopInLoop() {
	c.loop_.AssertInLoopGoroutine()
	if c.state_ == kConnecting {
		c.setState(kDisconnected)
		fd := c.removeAndResetChannel()
		c.retry(fd)
	}
}

func (c *Connector) connect() {
	fd := socket.CreateNonBlockOrDie()
	err := socket.Connect(fd, c.serverAddr_)
	switch err.(unix.Errno) {
	case 0:
		c.connecting(fd)
	case unix.EINPROGRESS:
		c.connecting(fd)
	case unix.EINTR:
		c.connecting(fd)
	case unix.EISCONN:
		c.connecting(fd)

	case unix.EAGAIN:
		c.retry(fd)
	case unix.EADDRINUSE:
		c.retry(fd)
	case unix.EADDRNOTAVAIL:
		c.retry(fd)
	case unix.ECONNREFUSED:
		c.retry(fd)
	case unix.ENETUNREACH:
		c.retry(fd)

	case unix.EACCES:
		Dlog.Printf("Connector.connect: connect error - [%s]\n", err.Error())
		unix.Close(fd)

	case unix.EPERM:
		Dlog.Printf("Connector.connect: connect error - [%s]\n", err.Error())
		unix.Close(fd)

	case unix.EAFNOSUPPORT:
		Dlog.Printf("Connector.connect: connect error - [%s]\n", err.Error())
		unix.Close(fd)

	case unix.EALREADY:
		Dlog.Printf("Connector.connect: connect error - [%s]\n", err.Error())
		unix.Close(fd)

	case unix.EBADF:
		Dlog.Printf("Connector.connect: connect error - [%s]\n", err.Error())
		unix.Close(fd)

	case unix.ENOTSOCK:
		Dlog.Printf("Connector.connect: connect error - [%s]\n", err.Error())
		unix.Close(fd)
	default:
		Dlog.Printf("Connector.connect: unexpected error - [%s]\n", err.Error())
		unix.Close(fd)
		break
	}
}

func (c *Connector) connecting(fd int) {
	c.setState(kConnecting)
	if c.channel_ != nil {
		log.Panicf("Connector.connecting: channel already set\n")
	}

	c.channel_ = NewChannel(c.loop_, int32(fd))
	c.channel_.SetWriteCallback(c.handleWrite)
	c.channel_.SetErrorCallback(c.handleError)
	c.channel_.EnableWriting()
}

func (c *Connector) setState(s int) {
	c.state_ = s
}

func (c *Connector) removeAndResetChannel() int {
	c.channel_.DisableAll()
	c.channel_.Remove()
	sockfd := c.channel_.Fd()
	c.loop_.QueueInLoop(c.resetChannel)
	return int(sockfd)
}

func (c *Connector) resetChannel() {
	c.channel_ = nil
}

func (c *Connector) retry(fd int) {
	unix.Close(fd)
	c.setState(kDisconnected)
	if atomic.LoadInt64(&c.connected_) == 1 {
		Dlog.Printf("Connector.retry(): retry connecting to [%s] in [%d] Millisecond\n", c.serverAddr_.String(), c.retryTime_)
		c.loop_.RunAfter(c.retryTime_, c.startInLoop)
		c.retryTime_ = min(c.retryTime_*2, MaxRetryTime)
	} else {
		Dlog.Printf("Connector.retry(): don't connect\n ")
	}
}

func (c *Connector) handleWrite() {
	Dlog.Printf("Connector.handleWrite() trace : state is %d\n", c.state_)
	if c.state_ == kConnecting {
		fd := c.removeAndResetChannel()
		opt, err := socket.GetSocketError(fd)
		if err != nil {
			Dlog.Printf("Connector.handleWrite(): get socket error failed \n")
		} else {
			if opt != 0 {
				Dlog.Printf("Connector.handleWrite() - SO_ERROR = %d\n", opt)
				c.retry(fd)
			} else if socket.IsSelfConnect(fd) {
				Dlog.Printf("Connector.handleWrite() - self Connect\n")
				c.retry(fd)
			} else {
				c.setState(kConnected)
				if atomic.LoadInt64(&c.connected_) == 1 {
					c.connectorNewConnectionCallback_(fd)
				} else {
					unix.Close(fd)
				}
			}
		}

	} else {
		if c.state_ != kDisconnected {
			log.Panicf("Connector.handleWrite(): state is wrong\n")
		}
	}
}

func (c *Connector) handleError() {
	Dlog.Printf("Connector.handleError() trace : state is %d\n", c.state_)
	if c.state_ == kConnecting {
		fd := c.removeAndResetChannel()
		c.resetChannel()
		opt, err := socket.GetSocketError(fd)
		if err != nil {
			Dlog.Printf("Connector.handleError(): get socket error failed \n")
		} else {
			Dlog.Printf("Connector.handleError() - SO_ERROR = %d\n", opt)
		}
		c.retry(fd)
	}
}
