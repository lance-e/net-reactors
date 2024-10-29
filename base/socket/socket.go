package socket

import (
	"log"
	"net/netip"

	"golang.org/x/sys/unix"
)

// create nonblock tcp socket fd
func CreateNonBlockOrDie() int {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, unix.IPPROTO_TCP)
	if err != nil || fd < 0 {
		log.Panicf("CreateNonBlockOrDie: create socketfd failed\n")

	}
	return fd
}
func BindOrDie(fd int, addr *netip.AddrPort) {
	if err := unix.Bind(fd, &unix.SockaddrInet4{
		Addr: addr.Addr().As4(),
		Port: int(addr.Port()),
	}); err != nil {
		log.Panicf("BindOrDie:bind socketfd:%d failed\n", fd)
	}
}

func ListenOrDie(fd int) {
	if err := unix.Listen(fd, unix.SOMAXCONN); err != nil {
		log.Panicf("ListenOrDie: listen socketfd:%d failed\n", fd)
	}
}

func Accept(fd int, client netip.AddrPort) {
	nfd, sa, err := unix.Accept4(fd, unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC)
	if err != nil {
		log.Printf("Accept:accept new connetction failed,error:%s\n", err.Error())
	}
	if nfd < 0 {
		//todo handle every case
	}
	return netip.AddrPort{}
}

func SetReuseAddr(fd int, isReuse bool) {
	opt := 0
	if isReuse {
		opt = 1
	}
	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, opt); err != nil {
		log.Printf("SetReuseAddr: set fd:%d reuse addr failed\n", fd)
	}
}

func SetReusePort(fd int, isReuse bool) {
	opt := 0
	if isReuse {
		opt = 1
	}
	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEPORT, opt); err != nil {
		log.Printf("SetReusePort: set fd:%d reuse port failed\n", fd)
	}
}
