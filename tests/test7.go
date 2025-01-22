package main

import (
	"fmt"
	"log"
	"net/netip"
	"os"
	"syscall"

	netreactors "github.com/lance-e/net-reactors"
)

func newConnetion(fd int, addr *netip.AddrPort) {
	fmt.Printf("newConnetion: accepted a new connection from %s\n", addr.Addr().String())
	n, err := syscall.Write(fd, []byte("How are you?\n"))
	if err != nil || n < 0 {
		log.Printf("write faild, error:%s\n ", err.Error())
	}
}

func main7() {
	fmt.Printf("main: pid :%d\n", os.Getpid())
	addr := netip.MustParseAddrPort("127.0.0.1:80")

	loop := netreactors.NewEventLoop()

	acceptor := netreactors.NewAcceptor(loop, &addr, false)
	acceptor.SetAcceptorNewConnectionCallback(newConnetion)
	acceptor.Listen()

	loop.Loop()
}
