package socket

import (
	"fmt"
	"log"
	"net/netip"
	"syscall"

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

func Accept4(fd int) (int, *netip.AddrPort) {
	nfd, sa, err := unix.Accept4(fd, unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC)
	if nfd < 0 && err != nil {
		switch err.(syscall.Errno) {
		case syscall.EAGAIN:
		case syscall.ECONNABORTED:
		case syscall.EINTR:
		case syscall.EPROTO:
		case syscall.EPERM:
		case syscall.EMFILE:
			//the above are expected error ,so ignore
		case syscall.EBADF:
			log.Fatalf("Accept4: unexpected error of accept ,error:%s\n", err.Error())
		case syscall.EFAULT:
			log.Fatalf("Accept4: unexpected error of accept ,error:%s\n", err.Error())
		case syscall.EINVAL:
			log.Fatalf("Accept4: unexpected error of accept ,error:%s\n", err.Error())
		case syscall.ENFILE:
			log.Fatalf("Accept4: unexpected error of accept ,error:%s\n", err.Error())
		case syscall.ENOBUFS:
			log.Fatalf("Accept4: unexpected error of accept ,error:%s\n", err.Error())
		case syscall.ENOMEM:
			log.Fatalf("Accept4: unexpected error of accept ,error:%s\n", err.Error())
		case syscall.ENOTSOCK:
			log.Fatalf("Accept4: unexpected error of accept ,error:%s\n", err.Error())
		case syscall.EOPNOTSUPP:
			log.Fatalf("Accept4: unexpected error of accept ,error:%s\n", err.Error())
		default:
			log.Fatalf("Accept4: unknown error of accept ,error:%s\n", err.Error())

		}
	}
	client := netip.AddrPort{}
	switch addr := sa.(type) {
	case *unix.SockaddrInet4:
		ip, ok := netip.AddrFromSlice(addr.Addr[:])
		if !ok {
			log.Printf("Accept4: parse ip address failed\n")
			break
		}
		client = netip.AddrPortFrom(ip, uint16(addr.Port))
	case *unix.SockaddrInet6:
		log.Printf("Accept4:don't handle ipv6\n")
		break
	case *unix.SockaddrUnix:
		log.Printf("Accept4:don't handle unix family\n")
		break
	default:
		log.Printf("Accept4: unknown socket address type\n")
	}

	return nfd, &client
}

func Connect(fd int, addr *netip.AddrPort) error {
	err := unix.Connect(fd, &unix.SockaddrInet4{
		Addr: addr.Addr().As4(),
		Port: int(addr.Port()),
	})
	return err
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

func GetLocalAddr(socketfd int) *netip.AddrPort {
	sa, err := unix.Getsockname(socketfd)
	if err != nil {
		log.Printf("GetLocalAddr: get socket name failed\n")
	}
	var addr netip.AddrPort
	switch a := sa.(type) {
	case *unix.SockaddrInet4:
		ip, ok := netip.AddrFromSlice(a.Addr[:])
		if !ok {
			log.Printf("GetLocalAddr: parse ip address failed\n")
			break
		}
		addr = netip.AddrPortFrom(ip, uint16(a.Port))
	default:
		log.Printf("GetLocalAddr: not support to handle other address family temporary\n")
	}
	return &addr
}

func GetPeerAddr(socketfd int) *netip.AddrPort {
	sa, err := unix.Getpeername(socketfd)
	if err != nil {
		log.Printf("GetPeerAddr: get socket name failed\n")
	}
	var addr netip.AddrPort
	switch a := sa.(type) {
	case *unix.SockaddrInet4:
		ip, ok := netip.AddrFromSlice(a.Addr[:])
		if !ok {
			log.Printf("GetPeerAddr: parse ip address failed\n")
			break
		}
		addr = netip.AddrPortFrom(ip, uint16(a.Port))
	default:
		log.Printf("GetPeerAddr: not support to handle other address family temporary\n")
	}
	return &addr
}

func IsSelfConnect(fd int) bool {
	local := GetLocalAddr(fd)
	peer := GetPeerAddr(fd)
	return local.Compare(*peer) == 0
}

func GetSocketError(fd int) (int, error) {
	opt, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_ERROR)
	if err != nil {
		fmt.Printf("GetSocketError: get soket option filed , error is %s\n", err.Error())
		return -1, err
	}
	return opt, nil
}

func ShutDownWrite(fd int) {
	err := unix.Shutdown(fd, unix.SHUT_WR)
	if err != nil {
		fmt.Printf("Shutdown: shut down write failed , error is %s\n", err.Error())
	}
}

func SetTcpNoDelay(fd int, on bool) {
	opt := 0
	if on {
		opt = 1
	}
	err := unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_NODELAY, opt)
	if err != nil {
		fmt.Printf("Socket.SetTcpNoDelay: set failed , error is [%s]\n", err.Error())
	}
}

func SetKeepAlive(fd int, on bool) {
	opt := 0
	if on {
		opt = 1
	}
	err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_KEEPALIVE, opt)
	if err != nil {
		fmt.Printf("Socket.SetKeepAlive: set failed , error is [%s]\n", err.Error())
	}
}
