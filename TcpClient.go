package netreactors

import (
	"fmt"
	"log"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lance-e/net-reactors/base/socket"
)

type TcpClient struct {
	loop_       *EventLoop
	name_       string
	connector_  *Connector
	retry_      int64 //atomic
	connected_  int64 //atomic
	mutex_      sync.Mutex
	connection_ *TcpConnection
	nextConnId_ int

	connectionCallback_    ConnectionCallback
	messageCallback_       MessageCallback
	writeCompleteCallback_ WriteCompleteCallback
}

// *************************
// public:
// *************************

func NewTcpClient(loop *EventLoop, serverAddr *netip.AddrPort, name string) (client *TcpClient) {
	client = &TcpClient{
		loop_:       loop,
		name_:       name,
		connector_:  NewConnector(loop, serverAddr),
		retry_:      0,
		connected_:  1,
		nextConnId_: 1,

		//default callback set:
		connectionCallback_:    func(tc *TcpConnection) {},
		messageCallback_:       func(tc *TcpConnection, b *Buffer, t time.Time) {},
		writeCompleteCallback_: nil,
	}
	client.connector_.SetConnectorNewConnectionCallback(client.newConnection)
	Dlog.Printf("NewTcpClient() trace: TcpClient [%s] - connector[%d]\n", client.Name(), &client.connector_)

	return
}

func (tc *TcpClient) Connect() {
	Dlog.Printf("TcpClient-[%s] connecting to [%s]\n", tc.Name(), tc.connector_.serverAddr_.String())
	atomic.StoreInt64(&tc.connected_, 1)
	tc.connector_.Start()
}

func (tc *TcpClient) Disconnect() {
	atomic.StoreInt64(&tc.connected_, 0)

	tc.mutex_.Lock()
	if tc.connection_ != nil {
		tc.connection_.Shutdown()
	}
	tc.mutex_.Unlock()
}

func (tc *TcpClient) Stop() {
	atomic.StoreInt64(&tc.connected_, 0)
	tc.connector_.Stop()
}

func (tc *TcpClient) GetLoop() *EventLoop {
	return tc.loop_
}

func (tc *TcpClient) Name() string {
	return tc.name_
}

func (tc *TcpClient) IsRetry() bool {
	return atomic.LoadInt64(&tc.retry_) == 1
}

func (tc *TcpClient) EnableRetry() {
	atomic.StoreInt64(&tc.retry_, 1)
}

func (tc *TcpClient) SetConnectionCallback(cb ConnectionCallback) {
	tc.connectionCallback_ = cb
}

func (tc *TcpClient) SetMessageCallback(cb MessageCallback) {
	tc.messageCallback_ = cb
}

func (tc *TcpClient) SetWriteCompleteCallback(cb WriteCompleteCallback) {
	tc.writeCompleteCallback_ = cb
}

// *************************
// private:
// *************************

func (tc *TcpClient) newConnection(fd int) {
	tc.loop_.AssertInLoopGoroutine()
	peerAddr := socket.GetPeerAddr(fd)
	connName := tc.Name() + fmt.Sprintf(":%s#%d", peerAddr.String(), tc.nextConnId_)
	tc.nextConnId_++

	conn := NewTcpConnection(tc.GetLoop(), connName, fd, socket.GetLocalAddr(fd), peerAddr)
	conn.SetConnectionCallback(tc.connectionCallback_)
	conn.SetMessageCallback(tc.messageCallback_)
	conn.SetWriteCompleteCallback(tc.writeCompleteCallback_)
	conn.SetCloseCallback(tc.removeConnection)

	tc.mutex_.Lock()
	tc.connection_ = conn
	tc.mutex_.Unlock()

	conn.ConnectEstablished()
}

func (tc *TcpClient) removeConnection(conn *TcpConnection) {
	tc.loop_.AssertInLoopGoroutine()
	if tc.loop_ != tc.GetLoop() {
		log.Panicf("removeConnection: the loop had changed\n")
	}

	tc.mutex_.Lock()
	if conn.Name() != tc.connection_.Name() {
		log.Panicf("removeConnection: the connection isn't same one\n")
	}
	tc.connection_ = nil //fixme: reset
	tc.mutex_.Unlock()

	tc.loop_.QueueInLoop(tc.connection_.ConnectDestroyed)
	if atomic.LoadInt64(&tc.retry_) == 1 && atomic.LoadInt64(&tc.connected_) == 1 {
		Dlog.Printf("TcpClient-[%s] reconnecting to [%s]\n", tc.Name(), tc.connector_.serverAddr_.String())
		tc.connector_.Restart()
	}

}
