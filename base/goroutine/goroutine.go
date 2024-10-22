package goroutine

import "github.com/petermattis/goid"

func GetGoid() int64 {
	return goid.Get()
}
