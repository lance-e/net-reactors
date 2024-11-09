package main

import (
	"fmt"
	netreactors "net-reactors"
	"net/netip"
	"os"
	"os/signal"
	"runtime/pprof"
	"time"
)

func onConnection(conn *netreactors.TcpConnection) {
	/* if conn.Connected() { */
	/* fmt.Printf("onConnection: new connection [%s] from %s \n", conn.Name(), conn.PeerAddr().String()) */
	/* } else { */
	/* fmt.Printf("onConnection: connection [%s] is down \n", conn.Name()) */
	/* } */
}

func onMessage(conn *netreactors.TcpConnection, buf *netreactors.Buffer, t time.Time) {
	// fmt.Printf("onMessage: received %d bytes from connection [%s] at time-[%s]\n", buf.ReadableBytes(), conn.Name(), t.Format("2006-01-02 15:04:05"))
	conn.Send(buf.RetrieveAllString())
}

func main() {
	f, err := os.Create("prof")
	if err != nil {
		panic(err)
	}
	pprof.StartCPUProfile(f)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		pprof.StopCPUProfile()
		os.Exit(0)
	}()
	fmt.Printf("main: pid %d\n", os.Getpid())

	addr := netip.MustParseAddrPort("127.0.0.1:80")
	loop := netreactors.NewEventLoop()

	server := netreactors.NewTcpServer(loop, &addr, "echoServer", 4)
	server.SetConnectionCallback(onConnection)
	server.SetMessageCallback(onMessage)
	// server.SetGoroutineNum(8)

	server.Start()

	loop.Loop()
}
