package main

import (
	"fmt"
	"log"
	netreactors "net-reactors"
	"net/netip"
	"os"
	"syscall"
)

func newConnetion(fd int, addr *netip.AddrPort) {
	fmt.Printf("newConnetion: accepted a new connection from %s\n", addr.Addr().String())
	n, err := syscall.Write(fd, []byte("How are you?\n"))
	if err != nil || n < 0 {
		log.Printf("write faild, error:%s\n ", err.Error())
	}
}

func main() {
	fmt.Printf("main: pid :%d\n", os.Getpid())
	addr := netip.MustParseAddrPort("127.0.0.1:80")

	loop := netreactors.NewEventLoop()

	acceptor := netreactors.NewAcceptor(loop, &addr, false)
	acceptor.SetNewConnectionCallback(newConnetion)
	acceptor.Listen()

	loop.Loop()
}
