package netinfo

import (
	"testing"
)

// Benchmark_Get-4      2000      890222 ns/op
// => roughly 1 ms

func Benchmark_Get(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Get()
	}
}
