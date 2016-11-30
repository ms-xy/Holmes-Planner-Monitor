// +build linux

// This package extends the idea found at https://github.com/capnm/sysinfo
package sysinfo

import (
	"testing"
)

// Result:
// Benchmark_UpdateSysinfo-4	10000000	       199 ns/op
// Benchmark_UpdateCores-4  	   20000	     94889 ns/op
// Benchmark_UpdateMeminfo-4	   10000	    156439 ns/op
// Benchmark_UpdateCpuinfo-4	    2000	    541624 ns/op

func Benchmark_UpdateSysinfo(b *testing.B) {
	si := &Sysinfo{}
	for i := 0; i < b.N; i++ {
		si.UpdateSysinfo()
	}
}

func Benchmark_UpdateCores(b *testing.B) {
	si := &Sysinfo{}
	for i := 0; i < b.N; i++ {
		si.UpdateCores()
	}
}

func Benchmark_UpdateMeminfo(b *testing.B) {
	si := &Sysinfo{}
	for i := 0; i < b.N; i++ {
		si.UpdateMeminfo()
	}
}

func Benchmark_UpdateCpuinfo(b *testing.B) {
	si := &Sysinfo{}
	for i := 0; i < b.N; i++ {
		si.UpdateCpuinfo()
	}
}
