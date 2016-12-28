// +build linux

// This package extends the idea found at https://github.com/capnm/sysinfo
package sysinfo

import (
	"testing"
)

// Result:
// Benchmark_UpdateSysinfo-4 	10000000	       194 ns/op
// Benchmark_UpdateCores-4   	   20000	     76628 ns/op
// Benchmark_UpdateMeminfo-4 	   10000	    110613 ns/op
// Benchmark_UpdateCpuinfo-4 	    5000	    448167 ns/op
// Benchmark_UpdateDiskinfo-4	    1000	   1125238 ns/op
// Benchmark_UpdateCpuUsage-4	   10000	    103789 ns/op

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

func Benchmark_UpdateDiskinfo(b *testing.B) {
	si := &Sysinfo{}
	for i := 0; i < b.N; i++ {
		si.UpdateDiskinfo()
	}
}

func Benchmark_UpdateCpuUsage(b *testing.B) {
	si := &Sysinfo{}
	for i := 0; i < b.N; i++ {
		si.UpdateCpuUsage()
	}
}
