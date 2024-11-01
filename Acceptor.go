package netreactors

import (
	"log"
	"net-reactors/base/socket"
	"net/netip"
	"time"

	"golang.org/x/sys/unix"
)

type Acceptor struct {
	loop_                  *EventLoop
	socketfd_              int
	acceptChannel_         *Channel
	newConnectionCallback_ NewConnectionCallback
	listening_             bool
}

// *************************
// public:
// *************************

func NewAcceptor(loop *EventLoop, listenAddr *netip.AddrPort, reusePort bool) (ac *Acceptor) {
	//set socketfd
	fd := socket.CreateNonBlockOrDie()
	socket.SetReuseAddr(fd, true)
	socket.SetReusePort(fd, reusePort)
	socket.BindOrDie(fd, listenAddr)

	//acceptor object
	ac = &Acceptor{
		loop_:                  loop,
		socketfd_:              fd,
		acceptChannel_:         NewChannel(loop, int32(fd)),
		newConnectionCallback_: nil,
		listening_:             false,
	}

	//set channel
	ac.acceptChannel_.SetReadCallback(ac.handleRead)
	return
}

func (a *Acceptor) SetNewConnectionCallback(cb NewConnectionCallback) {
	a.newConnectionCallback_ = cb
}

func (a *Acceptor) Listen() {
	a.loop_.AssertInLoopGoroutine()
	a.listening_ = true
	socket.ListenOrDie(a.socketfd_)
	a.acceptChannel_.EnableReading() //begin handle socketfd_'s read event
}

func (a *Acceptor) Listening() bool {
	return a.listening_
}

// *************************
// private:
// *************************

func (a *Acceptor) handleRead(time time.Time) {
	a.loop_.AssertInLoopGoroutine()
	connfd, addr := socket.Accept4(a.socketfd_)
	if connfd >= 0 {
		if a.newConnectionCallback_ != nil {
			a.newConnectionCallback_(connfd, addr)
		} else {
			unix.Close(connfd)
		}
	} else {
		log.Printf("Acceptor:handleRead accept new connection happened error\n")
		//todo handle the special error
		//if fd all use , here can do a special handle
	}
}
