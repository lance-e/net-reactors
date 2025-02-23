package main

import (
	"fmt"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"time"

	_ "net/http/pprof"

	netreactors "github.com/lance-e/net-reactors"
)

func onConnection(conn *netreactors.TcpConnection) {
	if conn.Connected() {
		fmt.Printf("onConnection: new connection [%s] from %s \n", conn.Name(), conn.PeerAddr().String())
	} else {
		fmt.Printf("onConnection: connection [%s] is down \n", conn.Name())
	}
}

func onMessage(conn *netreactors.TcpConnection, buf *netreactors.Buffer, t time.Time) {
	// fmt.Printf("onMessage: received %d bytes from connection [%s] at time-[%s]\n", buf.ReadableBytes(), conn.Name(), t.Format("2006-01-02 15:04:05"))
	conn.Send(buf.RetrieveAllString())

}

func main() {
	f, err := os.Create("pprof")
	if err != nil {
		panic(err)
	}
	runtime.SetBlockProfileRate(1)
	pprof.StartCPUProfile(f)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		pprof.StopCPUProfile()
		os.Exit(0)
	}()

	go func() {
		_ = http.ListenAndServe("0.0.0.0:6060", nil)
	}()

	addr := netip.MustParseAddrPort("127.0.0.1:80")
	loop := netreactors.NewEventLoop()

	server := netreactors.NewTcpServer(loop, &addr, "echoServer")
	server.SetConnectionCallback(onConnection)
	server.SetMessageCallback(onMessage)
	// server.SetGoroutineNum(runtime.NumCPU())

	server.Start()

	loop.Loop()
}
