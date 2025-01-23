package netreactors

import (
	"log"
	"net/netip"
	"time"

	"github.com/lance-e/net-reactors/base/socket"
	"golang.org/x/sys/unix"
)

const (
	kConnecting = iota
	kConnected
	kDisconnected
	kDisconnecting
)

type TcpConnection struct {
	loop_     *EventLoop
	name_     string
	state_    int
	socketfd_ int
	channel_  *Channel

	localAddr_ *netip.AddrPort
	peerAddr_  *netip.AddrPort

	connectionCallback_    ConnectionCallback
	messageCallback_       MessageCallback
	closeCallback_         CloseCallback
	highWaterMarkCallback_ HighWaterMarkCallback
	writeCompleteCallback_ WriteCompleteCallback

	inBuffer_  *Buffer
	outBuffer_ *Buffer
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
		inBuffer_:  NewBuffer(),
		outBuffer_: NewBuffer(),
	}

	conn.channel_.SetReadCallback(conn.handleRead)
	conn.channel_.SetWriteCallback(conn.handleWrite)
	conn.channel_.SetErrorCallback(conn.handleError)
	conn.channel_.SetCloseCallback(conn.handleClose)

	socket.SetKeepAlive(conn.socketfd_, true) //set tcp keep alive
	return
}

// send the 'message'
func (tc *TcpConnection) Send(message []byte) {
	if tc.state_ == kConnected {
		if tc.loop_.IsInLoopGoroutine() {
			tc.sendInLoop(message)
		} else {
			tc.loop_.RunInLoop(tc.bindSendInLoop(message))
		}
	}
}

// send the data from 'buf'
func (tc *TcpConnection) SendFromBuffer(buf *Buffer) {
	if tc.state_ == kConnected {
		if tc.loop_.IsInLoopGoroutine() {
			tc.sendInLoop(buf.RetrieveAllString())
		} else {
			tc.loop_.RunInLoop(tc.bindSendInLoop(buf.RetrieveAllString()))
		}
	}
}

// ShutDown
func (tc *TcpConnection) Shutdown() {
	if tc.state_ == kConnected {
		tc.setState(kDisconnecting)
		tc.loop_.RunInLoop(tc.shutdownInLoop)
	}
}

func (tc *TcpConnection) SetTcpNoDelay(on bool) {
	socket.SetTcpNoDelay(tc.socketfd_, on)
}

// set callback
func (tc *TcpConnection) SetConnectionCallback(cb ConnectionCallback) {
	tc.connectionCallback_ = cb
}
func (tc *TcpConnection) SetMessageCallback(cb MessageCallback) {
	tc.messageCallback_ = cb
}
func (tc *TcpConnection) SetCloseCallback(cb CloseCallback) {
	tc.closeCallback_ = cb
}
func (tc *TcpConnection) SetHighWaterMarkCallback(cb HighWaterMarkCallback) {
	tc.highWaterMarkCallback_ = cb
}
func (tc *TcpConnection) SetWriteCompleteCallback(cb WriteCompleteCallback) {
	tc.writeCompleteCallback_ = cb
}
func (tc *TcpConnection) BindWriteCompleteCallback() func() {
	return func() {
		tc.writeCompleteCallback_(tc)
	}
}

// called when TcpServer accepts a new connection
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

// called when TcpServer has removed me from it's map
func (tc *TcpConnection) ConnectDestroyed() {
	tc.loop_.AssertInLoopGoroutine()
	if tc.state_ == kConnected {
		tc.setState(kDisconnected)
		tc.channel_.DisableAll()
		tc.connectionCallback_(tc)
	}
	tc.channel_.Remove()
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

func (tc *TcpConnection) InBuffer() *Buffer {
	return tc.inBuffer_
}

func (tc *TcpConnection) OutBuffer() *Buffer {
	return tc.outBuffer_
}

// *************************
// private:
// *************************

func (tc *TcpConnection) handleRead(time time.Time) {
	tc.loop_.AssertInLoopGoroutine()
	var saveErrno error
	n := tc.inBuffer_.ReadFd(tc.socketfd_, &saveErrno)
	if n > 0 {
		tc.messageCallback_(tc, tc.inBuffer_, time) //get time by argument
	} else if n == 0 {
		tc.handleClose()
	} else {
		Dlog.Printf("TcpConnection.handleRead: read failed,err:%s\n", saveErrno.Error())
		tc.handleError()
	}
}

func (tc *TcpConnection) handleWrite() {
	tc.loop_.AssertInLoopGoroutine()
	if tc.channel_.IsWriting() {
		n, err := unix.Write(int(tc.channel_.Fd()), tc.outBuffer_.buffer_[tc.outBuffer_.readerIndex_:tc.outBuffer_.writerIndex_])
		if n >= 0 {
			tc.outBuffer_.Retrieve(n)
			if tc.outBuffer_.ReadableBytes() == 0 {
				tc.channel_.DisableWriting()
				if tc.writeCompleteCallback_ != nil {
					tc.loop_.QueueInLoop(tc.BindWriteCompleteCallback())
				}
			}

			if tc.state_ == kDisconnecting {
				tc.shutdownInLoop()
			}
		} else {
			Dlog.Printf("TcpConnection.handleWrite: write from outBuffer_ failed, error is %s\n", err.Error())
		}
	} else {
		Dlog.Printf("TcpConnection.handleWrite: connection is down ,no more writing\n")
	}
}

func (tc *TcpConnection) handleClose() {
	tc.loop_.AssertInLoopGoroutine()
	Dlog.Printf("TcpConnection:handleClose connection's state = %d\n", tc.state_)
	if tc.state_ != kConnected && tc.state_ != kDisconnecting {
		log.Panicf("TcpConnection:handleClose state isn't kConnected\n")
	}

	tc.channel_.DisableAll()

	tc.closeCallback_(tc)
}

func (tc *TcpConnection) handleError() {
	errno, err := socket.GetSocketError(tc.socketfd_)
	if err != nil {
		Dlog.Printf("TcpConnection.handleError: get socket errno failed\n")
	} else {
		Dlog.Printf("TcpConnection.handleError [%s] - SO_ERROR = %d\n", tc.name_, errno)
	}
}

func (tc *TcpConnection) bindSendInLoop(msg []byte) func() {
	return func() {
		tc.sendInLoop(msg)
	}
}

// run in loop goroutine
func (tc *TcpConnection) sendInLoop(msg []byte) {
	tc.loop_.AssertInLoopGoroutine()

	var n int
	var remaining = len(msg)
	faultError := false
	var err error
	if tc.state_ == kDisconnected {
		Dlog.Printf("disconnected , stop writing...\n")
		return
	}
	if !tc.channel_.IsWriting() && tc.outBuffer_.ReadableBytes() == 0 {
		n, err = unix.Write(int(tc.channel_.Fd()), msg)
		if n >= 0 {
			remaining = len(msg) - n
			if n < len(msg) {
				Dlog.Printf("I am going to write more data\n")
			} else if tc.writeCompleteCallback_ != nil {
				tc.loop_.QueueInLoop(tc.BindWriteCompleteCallback())
			}
		} else {
			n = 0
			if err != unix.EWOULDBLOCK {
				Dlog.Printf("TcpConnection.sendInLoop error [%s]\n", err.Error())
				if err == unix.EPIPE || err == unix.ECONNRESET {
					faultError = true
				}
			}
		}
	}

	if !faultError && remaining > 0 {
		//todo:
		//run highWaterMarkCallback_

		tc.outBuffer_.buffer_ = append(tc.outBuffer_.buffer_, msg[n:]...)
		if !tc.channel_.IsWriting() {
			tc.channel_.EnableWriting()
		}
	}
}
func (tc *TcpConnection) shutdownInLoop() {
	tc.loop_.AssertInLoopGoroutine()
	if !tc.channel_.IsWriting() {
		socket.ShutDownWrite(tc.socketfd_)
	}
}

func (tc *TcpConnection) setState(s int) {
	tc.state_ = s
}
