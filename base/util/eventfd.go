package util

import (
	"log"

	"golang.org/x/sys/unix"
)

func CreateEventFd() int {
	eventfd, err := unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC)
	if err != nil || eventfd < 0 {
		log.Panicln("CreateEventFd: create Eventfd failed")
	}
	return eventfd
}
