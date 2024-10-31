package main

import (
	"bytes"
	"fmt"
	netreactors "net-reactors"
	"net/netip"
	"os"
	"time"
)

func onConnection(conn *netreactors.TcpConnection) {
	if conn.Connected() {
		fmt.Printf("onConnection: new connection [%s] from %s \n", conn.Name(), conn.PeerAddr().String())
	} else {
		fmt.Printf("onConnection: connection [%s] is down \n", conn.Name())
	}
}

func onMessage(conn *netreactors.TcpConnection, buf *bytes.Buffer, t time.Time) {
	fmt.Printf("onMessage: received %d bytes from connection [%s] \n", buf.Len(), conn.Name())
}

func main8() {
	fmt.Printf("main: pid %d\n", os.Getpid())

	addr := netip.MustParseAddrPort("127.0.0.1:80")
	loop := netreactors.NewEventLoop()

	server := netreactors.NewTcpServer(loop, &addr, "test8_server")
	server.SetConnectionCallback(onConnection)
	server.SetMessageCallback(onMessage)

	server.Start()

	loop.Loop()
}
