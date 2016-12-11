// +build linux

// This package extends the idea found at https://github.com/capnm/sysinfo
package sysinfo

// #include <linux/sysinfo.h>
import "C"

import (
	"errors"
	"github.com/c9s/goprocinfo/linux"
	"github.com/ms-xy/Holmes-Planner-Monitor/go/client/diskinfo"
	"io/ioutil"
	"syscall"
)

// Pretty fast (200ns). However, inaccuracy of memory values renders it useless
// for UpdateMeminfo. Load values and uptime seem to be accurate though.
func (this *Sysinfo) UpdateSysinfo() error {

	si := &syscall.Sysinfo_t{}

	err := syscall.Sysinfo(si)
	if err != nil {
		return err
	}

	shift := float64(1 << C.SI_LOAD_SHIFT)
	this.System.Load[0] = float64(si.Loads[0]) / shift
	this.System.Load[1] = float64(si.Loads[1]) / shift
	this.System.Load[2] = float64(si.Loads[2]) / shift

	// this.System.Uptime = time.Duration(si.Uptime) * time.Second
	this.System.Uptime = si.Uptime

	return nil
}

// Acceptable speed (90ms). As long as we don't ship any other CPU information,
// The little information yield of this one is also acceptable.
func (this *Sysinfo) UpdateCores() error {
	data, err := ioutil.ReadFile("/proc/cpuinfo")
	if err != nil {
		return errors.New("Error reading /proc/cpuinfo: " + err.Error())
	}

	lines := splitLines(data)

	var ncores = 0

	for _, line := range lines {
		switch line[0] {
		case 'p':
			if startsWith(line, []byte("processor")) {
				ncores++
			}
		}
	}

	this.Cpu.Cores = ncores

	return nil
}

func splitLines(buf []byte) [][]byte {
	lines := [][]byte{}
	s, e := 0, 0

	for ; e < len(buf); e++ {
		if buf[e] == '\n' {
			if e-s > 1 {
				lines = append(lines, buf[s:e])
			}
			s = e + 1
		}
	}

	if e-s > 0 {
		lines = append(lines, buf[s:e])
	}

	return lines
}

func startsWith(line []byte, expr []byte) bool {
	for i := 0; i < len(expr); i++ {
		if line[i] != expr[i] {
			return false
		}
	}
	return true
}

// Acceptable speed (150ms). Yields a lot of information.
// Parsing speed could potentially be increased a bit by dropping the
// framework and writing a parser. (The lib uses regular expressions even
// though the simplicity of the file entries does not require them)
func (this *Sysinfo) UpdateMeminfo() error {
	mi, err := linux.ReadMemInfo("/proc/meminfo")
	if err != nil {
		return err
	}

	// see htop source, e.g. at
	// https://github.com/hishamhm/htop/blob/master/linux/Platform.c
	// as reference for this calculation
	// TODO: find reason why MemTotal value in htop seem to be way off the
	// MemTotal reported in /proc/meminfo (7687MB vs 8.1GiB, not even explainable
	// using a different scaling, 1000 would yield less ... values in
	// /proc/meminfo are hardcoded with factor 1024 (2<<10) though)
	// (Same applies to all other values too as it seems all a bit lower in htop)
	total := mi.MemTotal
	free := mi.MemFree
	cached := mi.Cached + mi.SReclaimable - mi.Shmem
	available := free + mi.Buffers + cached
	used := total - available

	this.Ram.Total = total * 1024
	this.Ram.Free = free * 1024
	this.Ram.Available = available * 1024
	this.Ram.Used = used * 1024

	this.Swap.Total = mi.SwapTotal * 1024
	this.Swap.Free = mi.SwapFree * 1024
	this.Swap.Available = mi.SwapFree * 1024
	this.Swap.Used = (mi.SwapTotal - mi.SwapFree) * 1024

	return nil
}

// For some reason this function requires 540ms to complete. That is a bit too
// much for the information we need (just one int ...).
// Reason probably is all the reallocations paired with inefficient conversions.
// (Inefficient use of append on a list of structs)
func (this *Sysinfo) UpdateCpuinfo() error {
	ci, err := linux.ReadCPUInfo("/proc/cpuinfo")
	if err != nil {
		return err
	}

	this.Cpu.Cores = ci.NumCPU()

	return nil
}

// Update the harddrive information of the system.
func (this *Sysinfo) UpdateDiskinfo() error {
	harddrives, err := diskinfo.Get()
	if err != nil {
		return err
	}

	this.Harddrives = harddrives

	return nil
}
