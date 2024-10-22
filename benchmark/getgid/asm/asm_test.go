package benchmark

import (
	"testing"

	"github.com/petermattis/goid"
)

func BenchmarkBuffWithPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		goid.Get()
	}
}
