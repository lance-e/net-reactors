package main

import (
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

func onMessage(conn *netreactors.TcpConnection, buf *netreactors.Buffer, t time.Time) {
	fmt.Printf("onMessage: received %d bytes from connection [%s] at time-[%s]\n", buf.ReadableBytes(), conn.Name(), t.Format("2006-01-02 15:04:05"))
	conn.Send(buf.RetrieveAllString())
}

func main() {
	fmt.Printf("main: pid %d\n", os.Getpid())

	addr := netip.MustParseAddrPort("127.0.0.1:80")
	loop := netreactors.NewEventLoop()

	server := netreactors.NewTcpServer(loop, &addr, "echoServer")
	server.SetConnectionCallback(onConnection)
	server.SetMessageCallback(onMessage)

	server.Start()

	loop.Loop()
}
