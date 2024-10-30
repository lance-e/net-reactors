package netreactors

import "net/netip"

type TcpServer struct {
	loop_               *EventLoop
	name_               string
	acceptor_           *Acceptor
	started_            bool
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
		started_:            false,
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

func (t *TcpServer) newConnection(fd int, addr *netip.AddrPort) {
	t.loop_.AssertInLoopGoroutine()
	//to create new connection

}
