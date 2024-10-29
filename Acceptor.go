package netreactors

import (
	"net-reactors/base/socket"
	"net/netip"
)

type NewConnectionCallback func(int, *InetAddress)

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

func (a *Acceptor) NewAcceptor(loop *EventLoop, listenAddr *netip.AddrPort, reusePort bool) (ac Acceptor) {
	//set socketfd
	fd := socket.CreateNonBlockOrDie()
	socket.SetReuseAddr(fd, true)
	socket.SetReusePort(fd, reusePort)
	socket.BindOrDie(fd, listenAddr)

	//acceptor object
	ac = Acceptor{
		loop_:                  loop,
		socketfd_:              fd,
		acceptChannel_:         NewChannel(loop, int32(fd)),
		newConnectionCallback_: nil,
		listening_:             false,
	}

	//set channel
	ac.acceptChannel_.SetReadCallback(a.handleRead)

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

func (a *Acceptor) handleRead() {

}
