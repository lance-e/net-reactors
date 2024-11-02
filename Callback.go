package netreactors

import (
	"net/netip"
	"time"
)

type (
	EventCallback                  func()
	ReadEventCallback              func(time.Time)
	TimerCallback                  func()
	Functor                        func()
	GoroutineCallback              func(*EventLoop)
	AcceptorNewConnectionCallback  func(int, *netip.AddrPort)
	ConnectorNewConnectionCallback func(int)
	//TcpConnection
	ConnectionCallback    func(*TcpConnection)
	CloseCallback         func(*TcpConnection)
	WriteCompleteCallback func(*TcpConnection)
	MessageCallback       func(*TcpConnection, *Buffer, time.Time)
	HighWaterMarkCallback func(*TcpConnection, int)
)
