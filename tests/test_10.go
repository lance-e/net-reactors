package main

import (
	"fmt"
	netreactors "net-reactors"
	"net/netip"
	"os"
	"strconv"
	"time"
)

var (
	msg1      []byte
	msg2      []byte
	sleeptime time.Duration
)

func onConnection10(conn *netreactors.TcpConnection) {
	if conn.Connected() {
		fmt.Printf("onConnection(): new connection [%s] form %s\n", conn.Name(), conn.PeerAddr().String())
		if sleeptime != 0 {
			time.Sleep(sleeptime)
		}
		conn.Send(msg1)
		conn.Send(msg2)
		conn.Shutdown()
	} else {
		fmt.Printf("onConnection(): connection [%s] is down\n", conn.Name())
	}

}
func onMessage10(conn *netreactors.TcpConnection, buf *netreactors.Buffer, t time.Time) {
	fmt.Printf("onMessage: received %d bytes from connection [%s] at time-[%s]\n", buf.ReadableBytes(), conn.Name(), t.Format("2006-01-02 15:04:05"))
	buf.RetrieveAll()
}
func main10() {
	fmt.Printf("main: pid %d\n", os.Getpid())

	len1 := 100
	len2 := 200

	if len(os.Args) > 2 {
		len1, _ = strconv.Atoi(os.Args[1])
		len2, _ = strconv.Atoi(os.Args[2])
	}

	sleeptime = 5 * time.Second

	msg1 = make([]byte, len1)
	msg2 = make([]byte, len2)

	for i, _ := range msg1 {
		msg1[i] = 'A'
	}

	for i, _ := range msg2 {
		msg2[i] = 'B'
	}

	addr := netip.MustParseAddrPort("127.0.0.1:80")
	loop := netreactors.NewEventLoop()

	server := netreactors.NewTcpServer(loop, &addr, "test10_server")
	server.SetConnectionCallback(onConnection10)
	server.SetMessageCallback(onMessage10)

	server.Start()

	loop.Loop()
}
