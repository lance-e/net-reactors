package netreactors

import (
	"bytes"
	"fmt"
	"log"
	"net/netip"
	"time"

	"golang.org/x/sys/unix"
)

const (
	kConnecting = iota
	kConnected
)

type TcpConnection struct {
	loop_               *EventLoop
	name_               string
	state_              int
	socketfd_           int
	channel_            *Channel
	localAddr_          *netip.AddrPort
	peerAddr_           *netip.AddrPort
	connectionCallback_ ConnectionCallback
	messageCallback_    MessageCallback
}

// *************************
// public:
// *************************
func NewTcpConnection(loop *EventLoop, name string, fd int, localAddr *netip.AddrPort, peerAddr *netip.AddrPort) (conn *TcpConnection) {
	conn = &TcpConnection{
		loop_:      loop,
		name_:      name,
		state_:     kConnecting,
		socketfd_:  fd,
		channel_:   NewChannel(loop, int32(fd)),
		localAddr_: localAddr,
		peerAddr_:  peerAddr,
	}

	conn.channel_.SetReadCallback(conn.handleRead)
	conn.channel_.SetWriteCallback(conn.handleWrite)
	conn.channel_.SetErrorCallback(conn.handleError)
	conn.channel_.SetCloseCallback(conn.handleClose)

	return
}

func (tc *TcpConnection) SetConnectionCallback(cb ConnectionCallback) {
	tc.connectionCallback_ = cb
}

func (tc *TcpConnection) SetMessageCallback(cb MessageCallback) {
	tc.messageCallback_ = cb
}

func (tc *TcpConnection) ConnectEstablished() {
	tc.loop_.AssertInLoopGoroutine()
	if tc.state_ != kConnecting {
		log.Panicf("TcpConnection's state not kConnecting\n")
	}
	tc.setState(kConnected)
	//begin handle socketfd's readable event
	tc.channel_.EnableReading()

	//callback
	tc.connectionCallback_(tc)
}

func (tc *TcpConnection) Connected() bool {
	return tc.state_ == kConnected
}

func (tc *TcpConnection) Name() string {
	return tc.name_
}

func (tc *TcpConnection) GetLoop() *EventLoop {
	return tc.loop_
}

func (tc *TcpConnection) LocalAddr() *netip.AddrPort {
	return tc.localAddr_
}
func (tc *TcpConnection) PeerAddr() *netip.AddrPort {
	return tc.peerAddr_
}

// *************************
// private:
// *************************

/* func (tc *TcpConnection) bindHandleRead() func() { */
/*  */
/* } */
func (tc *TcpConnection) handleRead() {
	tc.loop_.AssertInLoopGoroutine()

	buf := make([]byte, 65535)
	n, err := unix.Read(int(tc.channel_.Fd()), buf)
	if err != nil || n < 0 {
		log.Printf("TcpConnection.handleRead: read failed,err:%s\n", err.Error())
		tc.handleError()
	} else if n == 0 {
		tc.handleClose()
	} else {
		tc.messageCallback_(tc, bytes.NewBuffer(buf[:n]), time.Now()) //get time by argument
	}
}

func (tc *TcpConnection) handleWrite() {
	fmt.Printf("handleWrite\n")
}

func (tc *TcpConnection) handleClose() {
	fmt.Printf("handleClose\n")
}

func (tc *TcpConnection) handleError() {
	fmt.Printf("handleClose\n")
}

func (tc *TcpConnection) setState(s int) {
	tc.state_ = s
}
