package main

import (
	"fmt"
	"net/netip"
	"os"
	"time"

	netreactors "github.com/lance-e/net-reactors"
)

func onConnection8(conn *netreactors.TcpConnection) {
	if conn.Connected() {
		fmt.Printf("onConnection: new connection [%s] from %s \n", conn.Name(), conn.PeerAddr().String())
	} else {
		fmt.Printf("onConnection: connection [%s] is down \n", conn.Name())
	}
}

func onMessage8(conn *netreactors.TcpConnection, buf *netreactors.Buffer, t time.Time) {
	fmt.Printf("onMessage: received %d bytes from connection [%s] at time-[%s]\n", buf.ReadableBytes(), conn.Name(), t.Format("2006-01-02 15:04:05"))
	fmt.Printf("onMessage: [%s]\n", buf.RetrieveAllString())
}

func main8() {
	fmt.Printf("main: pid %d\n", os.Getpid())

	addr := netip.MustParseAddrPort("127.0.0.1:80")
	loop := netreactors.NewEventLoop()

	server := netreactors.NewTcpServer(loop, &addr, "test8_server")
	server.SetConnectionCallback(onConnection8)
	server.SetMessageCallback(onMessage8)

	server.Start()

	loop.Loop()
}
