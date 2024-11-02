package main

import (
	"fmt"
	netreactors "net-reactors"
	"net/netip"
	"os"
	"strings"
	"time"
)

var msg11 []byte

func onConnection11(conn *netreactors.TcpConnection) {
	if conn.Connected() {
		fmt.Printf("onConnection(): new connection [%s] form %s\n", conn.Name(), conn.PeerAddr().String())

		conn.Send(msg11)
	} else {
		fmt.Printf("onConnection(): connection [%s] is down\n", conn.Name())
	}
}

func onWrtiecomplete11(conn *netreactors.TcpConnection) {
	conn.Send(msg11)
}
func onMessage11(conn *netreactors.TcpConnection, buf *netreactors.Buffer, t time.Time) {
	fmt.Printf("onMessage: received %d bytes from connection [%s] at time-[%s]\n", buf.ReadableBytes(), conn.Name(), t.Format("2006-01-02 15:04:05"))
	buf.RetrieveAll()
}
func main11() {
	fmt.Printf("main: pid %d\n", os.Getpid())

	var line strings.Builder
	for i := 33; i < 127; i++ {
		line.WriteByte(byte(i))
	}
	line.WriteString(line.String()) //  相当于 C++ 的 line += line

	var message strings.Builder
	lineStr := line.String()
	for i := 0; i < 127-33; i++ {
		message.WriteString(lineStr[i:min(i+72, len(lineStr))]) //  相当于 C++ 的 line.substr(i, 72)
		message.WriteByte('\n')
	}

	msg11 = []byte(message.String())

	addr := netip.MustParseAddrPort("127.0.0.1:80")
	loop := netreactors.NewEventLoop()

	server := netreactors.NewTcpServer(loop, &addr, "test11_server")
	server.SetConnectionCallback(onConnection11)
	server.SetMessageCallback(onMessage11)
	server.SetWriteCompleteCallback(onWrtiecomplete11)

	server.Start()

	loop.Loop()
}
