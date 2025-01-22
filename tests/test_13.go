package main

import (
	"fmt"
	"net/netip"
	"time"

	netreactors "github.com/lance-e/net-reactors"
)

var hello = "hello!!!"

func onConnection13(conn *netreactors.TcpConnection) {
	if conn.Connected() {
		fmt.Printf("onConnection13(): new connection [%s] form %s\n", conn.Name(), conn.PeerAddr().String())

		conn.Send([]byte(hello))
	} else {
		fmt.Printf("onConnection13(): connection [%s] is down\n", conn.Name())
	}
}

func onWrtiecomplete13(conn *netreactors.TcpConnection) {
}
func onMessage13(conn *netreactors.TcpConnection, buf *netreactors.Buffer, t time.Time) {
	fmt.Printf("onMessage13: received %d bytes from connection [%s] at time-[%s]\n", buf.ReadableBytes(), conn.Name(), t.Format("2006-01-02 15:04:05"))
	fmt.Printf("onMessage13(): [%s]\n", buf.RetrieveAllString())
}
func main13() {

	addr := netip.MustParseAddrPort("127.0.0.1:80")
	loop := netreactors.NewEventLoop()

	client := netreactors.NewTcpClient(loop, &addr, "client13")
	client.SetConnectionCallback(onConnection13)
	client.SetMessageCallback(onMessage13)
	// client.SetWriteCompleteCallback(onWrtiecomplete11)
	client.EnableRetry()
	client.Connect()
	loop.Loop()
}
