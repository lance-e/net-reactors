package benchmark

import (
	"bytes"
	"errors"
	"runtime"
	"strconv"
	"testing"
)

func GetGoroutineId() (int64, error) {
	var goroutineSpace = []byte("goroutine ")
	bs := make([]byte, 128)
	bs = bs[:runtime.Stack(bs, false)]
	bs = bytes.TrimPrefix(bs, goroutineSpace)
	i := bytes.IndexByte(bs, ' ')
	if i < 0 {
		return -1, errors.New("get goroutine id failed")
	}
	return strconv.ParseInt(string(bs[:i]), 10, 64)

}

func BenchmarkBufferWithPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GetGoroutineId()
	}
}
