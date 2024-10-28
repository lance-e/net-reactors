package util

import (
	"log"
	"time"

	"golang.org/x/sys/unix"
)

// create
func CreateTimerFd() int {
	tfd, err := unix.TimerfdCreate(unix.CLOCK_MONOTONIC, unix.TFD_NONBLOCK|unix.TFD_CLOEXEC)
	if err != nil || tfd < 0 {
		log.Panicf("CreateTimerFd: TimerfdCreate failed\n")
	}
	return tfd
}

// read
func ReadTimerfd(fd int, now time.Time) {
	//read timefd
	var buf [8]byte
	n, err := unix.Read(fd, buf[:])
	howmany := uint64(buf[0]) | uint64(buf[1])<<8 | uint64(buf[2])<<16 | uint64(buf[3])<<24 | uint64(buf[4])<<32 | uint64(buf[5])<<40 | uint64(buf[6])<<48 | uint64(buf[7])<<56
	log.Printf("tq.HandleRead expirated %d times at now:%s\n ", howmany, now.Format("2006-01-02 15:04:05"))
	if err != nil || n != 8 {
		log.Printf("tq.HandleRead: read timerfd failed, read %d bytes instead of 8 bytes\n", n)
	}
}

// reset
func ResetTimerfd(fd int, expiration time.Time) {
	newValue := unix.ItimerSpec{}
	oldValue := unix.ItimerSpec{}
	newValue.Value = howMuchTimeFromNow(expiration)
	err := unix.TimerfdSettime(fd, 0, &newValue, &oldValue)
	if err != nil {
		log.Printf("ResetTimerfd :tiemrfd set failed\n")
	}
}

func howMuchTimeFromNow(when time.Time) unix.Timespec {
	microseconds := (when.UnixNano() / int64(time.Microsecond)) - (time.Now().UnixNano() / int64(time.Microsecond))
	if microseconds < 100 {
		microseconds = 100
	}
	return unix.Timespec{
		Sec:  microseconds / 1000000, //kmicroseconds per second
		Nsec: (microseconds % 1000000) * 1000,
	}
}
