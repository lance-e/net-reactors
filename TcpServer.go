package netreactors

import (
	"fmt"
	"log"
	"net-reactors/base/socket"
	"net/netip"
	"sync/atomic"
	"time"
)

type TcpServer struct {
	loop_          *EventLoop
	name_          string
	acceptor_      *Acceptor
	started_       int64 //atomic
	nextConnId_    int
	connections_   map[string]*TcpConnection
	goroutinePool_ *EventLoopGoroutinePool

	connectionCallback_    ConnectionCallback
	messageCallback_       MessageCallback
	writeCompleteCallback_ WriteCompleteCallback
	goroutineCallback_     GoroutineCallback
}

func NewTcpServer(loop *EventLoop, addr *netip.AddrPort, name string) (server *TcpServer) {
	server = &TcpServer{
		loop_:          loop,
		name_:          name,
		acceptor_:      NewAcceptor(loop, addr, false),
		started_:       0,
		nextConnId_:    1,
		connections_:   make(map[string]*TcpConnection, 0),
		goroutinePool_: NewEventLoopGoroutinePool(loop),

		connectionCallback_:    func(tc *TcpConnection) {},
		messageCallback_:       func(tc *TcpConnection, b *Buffer, t time.Time) {},
		writeCompleteCallback_: nil,
		goroutineCallback_:     nil,
	}

	//set callback when new connection coming
	server.acceptor_.SetAcceptorNewConnectionCallback(server.newConnection)
	return
}

// *************************
// public:
// *************************

func (t *TcpServer) Start() {
	if atomic.CompareAndSwapInt64(&t.started_, 0, 1) {
		//goroutine pool
		t.goroutinePool_.Start(t.goroutineCallback_)

		if t.acceptor_.listening_ {
			log.Panicf("tcpserver's acceptor is listening\n")
		}

		t.loop_.RunInLoop(t.acceptor_.Listen)
	}
}

func (t *TcpServer) SetGoroutineNum(num int) {
	if num < 0 {
		log.Panicf("TcpServer.SetGoroutineNum: wrong goroutine number \n")
	}
	t.goroutinePool_.SetGoroutineNum(num)
}

func (t *TcpServer) SetConnectionCallback(cb ConnectionCallback) {
	t.connectionCallback_ = cb
}
func (t *TcpServer) SetMessageCallback(cb MessageCallback) {
	t.messageCallback_ = cb
}
func (t *TcpServer) SetWriteCompleteCallback(cb WriteCompleteCallback) {
	t.writeCompleteCallback_ = cb
}
func (t *TcpServer) SetGoroutineCallback(cb GoroutineCallback) {
	t.goroutineCallback_ = cb
}

// *************************
// private:
// *************************

func (t *TcpServer) newConnection(fd int, peerAddr *netip.AddrPort) {
	t.loop_.AssertInLoopGoroutine()

	//one loop per goroutine!
	ioLoop := t.goroutinePool_.GetNextLoop()

	//to create new connection
	connName := fmt.Sprintf(t.name_+"#%d", t.nextConnId_)
	t.nextConnId_++
	log.Printf("TcpServer:newConnection [%s] - new connection [%s] from %s\n", t.name_, connName, peerAddr.String())

	//create new connection and set it's attribute
	conn := NewTcpConnection(ioLoop, connName, fd, socket.GetLocalAddr(fd), peerAddr)
	t.connections_[connName] = conn
	conn.SetConnectionCallback(t.connectionCallback_)
	conn.SetMessageCallback(t.messageCallback_)
	conn.SetCloseCallback(t.removeConnection)
	conn.SetWriteCompleteCallback(t.writeCompleteCallback_)

	//here is ioLoop to run:
	ioLoop.RunInLoop(conn.ConnectEstablished)
}

func (t *TcpServer) removeConnection(conn *TcpConnection) {
	t.loop_.RunInLoop(t.bindRemoveConnection(conn))
}

func (t *TcpServer) bindRemoveConnection(conn *TcpConnection) func() {
	return func() {
		t.removeConnectionInLoop(conn)
	}
}
func (t *TcpServer) removeConnectionInLoop(conn *TcpConnection) {
	t.loop_.AssertInLoopGoroutine()
	log.Printf("TcpServer:removeConnectionInLoop [%s] - connection [%s]\n", t.name_, conn.Name())
	delete(t.connections_, conn.Name())
	ioLoop := conn.GetLoop()
	ioLoop.QueueInLoop(conn.ConnectDestroyed)
}
