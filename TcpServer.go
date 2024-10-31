package netreactors

import (
	"fmt"
	"log"
	"net-reactors/base/socket"
	"net/netip"
	"sync/atomic"
)

type TcpServer struct {
	loop_               *EventLoop
	name_               string
	acceptor_           *Acceptor
	started_            int64 //atomic
	connectionCallback_ ConnectionCallback
	messageCallback_    MessageCallback
	nextConnId          int
	connections_        map[string]*TcpConnection
}

func NewTcpServer(loop *EventLoop, addr *netip.AddrPort, name string) (server *TcpServer) {
	server = &TcpServer{
		loop_:               loop,
		name_:               name,
		acceptor_:           NewAcceptor(loop, addr, false),
		started_:            0,
		connectionCallback_: nil,
		messageCallback_:    nil,
		nextConnId:          1,
		connections_:        make(map[string]*TcpConnection, 0),
	}

	//set callback when new connection coming
	server.acceptor_.SetNewConnectionCallback(server.newConnection)
	return
}

// *************************
// public:
// *************************

func (t *TcpServer) Start() {
	if atomic.CompareAndSwapInt64(&t.started_, 0, 1) {
		//goroutine pool

		if t.acceptor_.listening_ {
			log.Panicf("tcpserver's acceptor is listening\n")
		}

		t.loop_.RunInLoop(t.acceptor_.Listen)
	}
}

func (t *TcpServer) SetConnectionCallback(cb ConnectionCallback) {
	t.connectionCallback_ = cb
}

func (t *TcpServer) SetMessageCallback(cb MessageCallback) {
	t.messageCallback_ = cb
}

// *************************
// private:
// *************************

func (t *TcpServer) newConnection(fd int, peerAddr *netip.AddrPort) {
	t.loop_.AssertInLoopGoroutine()
	//to create new connection
	connName := fmt.Sprintf(t.name_+"#%d", t.nextConnId)
	t.nextConnId++
	log.Printf("TcpServer:newConnection [%s] - new connection [%s] from %s\n", t.name_, connName, peerAddr.String())

	//create new connection and set it's attribute
	conn := NewTcpConnection(t.loop_, connName, fd, socket.GetLocalAddr(fd), peerAddr)
	t.connections_[t.name_] = conn
	conn.SetConnectionCallback(t.connectionCallback_)
	conn.SetMessageCallback(t.messageCallback_)
	conn.ConnectEstablished()
}
