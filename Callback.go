package netreactors

import (
	"bytes"
	"net/netip"
	"time"
)

type (
	EventCallback         func()
	TimerCallback         func()
	Functor               func()
	GoroutineCallback     func(*EventLoop)
	NewConnectionCallback func(int, *netip.AddrPort)
	//TcpConnection
	ConnectionCallback    func(*TcpConnection)
	CloseCallback         func(*TcpConnection)
	WriteCompleteCallback func(*TcpConnection)
	MessageCallback       func(*TcpConnection, *bytes.Buffer, time.Time)
)
