package server

import (
	// "github.com/ms-xy/Holmes-Planner-Monitor/go/message"
	"fmt"
	"net"
	"testing"
)

// TODO: remove this test stuff
var (
	bytemap18key  [18]byte
	uint64map3key [3]uint64
	stringmapkey  string
	bytemap32key  [32]byte
	uint64map4key [4]uint64
	bytemap18     = make(map[[18]byte]int, 1)
	uint64map3    = make(map[[3]uint64]int, 1)
	stringmap     = make(map[string]int, 1)
	bytemap32     = make(map[[32]byte]int, 1)
	uint64map4    = make(map[[4]uint64]int, 1)
	addr, _       = net.ResolveUDPAddr("udp", "127.0.0.1:9016")
	addrstr       = "127.0.0.1:9016"
)

// -------------------------------------------------------------------------- //
// benchmark key generation

func Benchmark_Bytemap18_Keygeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bytemap18key = addr2bytemap18key(addr.IP, addr.Port)
	}
}

func Benchmark_Uint64map3_Keygeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		uint64map3key = addr2uint64map3key(addr.IP, addr.Port)
	}
}

func Benchmark_Stringmap_Addr2String_Keygeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		stringmapkey = addr.String()
	}
}

func Benchmark_Bytemap32_Keygeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bytemap32key = str2bytemap32key(addrstr)
	}
}

func Benchmark_Uint64map4_Keygeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		uint64map4key = str2uint64map4key(addrstr)
	}
}

// -------------------------------------------------------------------------- //
// benchmark insert

func Benchmark_Bytemap18_Insert(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bytemap18[bytemap18key] = i
	}
}

func Benchmark_Uint64map3_Insert(b *testing.B) {
	for i := 0; i < b.N; i++ {
		uint64map3[uint64map3key] = i
	}
}

func Benchmark_Stringmap_Insert(b *testing.B) {
	for i := 0; i < b.N; i++ {
		stringmap[stringmapkey] = i
	}
}

func Benchmark_Bytemap32_Insert(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bytemap32[bytemap32key] = i
	}
}

func Benchmark_Uint64map4_Insert(b *testing.B) {
	for i := 0; i < b.N; i++ {
		uint64map4[uint64map4key] = i
	}
}

// -------------------------------------------------------------------------- //
// benchmark lookup

func Benchmark_Bytemap18_Lookup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, exists := bytemap18[bytemap18key]; !exists {
			fmt.Println("whoop error")
		}
	}
}

func Benchmark_Uint64map3_Lookup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, exists := uint64map3[uint64map3key]; !exists {
			fmt.Println("whoop error")
		}
	}
}

func Benchmark_Stringmap_Lookup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, exists := stringmap[stringmapkey]; !exists {
			fmt.Println("whoop error")
		}
	}
}

func Benchmark_Bytemap32_Lookup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, exists := bytemap32[bytemap32key]; !exists {
			fmt.Println("whoop error")
		}
	}
}

func Benchmark_Uint64map4_Lookup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, exists := uint64map4[uint64map4key]; !exists {
			fmt.Println("whoop error")
		}
	}
}
