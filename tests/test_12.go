package main

import (
	"fmt"
	"net/netip"

	netreactors "github.com/lance-e/net-reactors"
)

var loop12 *netreactors.EventLoop

func connectCallback12(fd int) {
	fmt.Printf("-----connected------\n")
	loop12.Quit()
}
func main12() {
	loop := netreactors.NewEventLoop()
	loop12 = loop

	addr := netip.MustParseAddrPort("127.0.0.1:80")
	connector := netreactors.NewConnector(loop, &addr)
	connector.SetConnectorNewConnectionCallback(connectCallback12)
	connector.Start()

	loop.Loop()
}
