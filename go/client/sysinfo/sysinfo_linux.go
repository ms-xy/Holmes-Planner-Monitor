// +build linux

// This package extends the idea found at https://github.com/capnm/sysinfo
package sysinfo

// #include <linux/sysinfo.h>
import "C"

import (
	"errors"
	goprocinfo "github.com/c9s/goprocinfo/linux"
	"github.com/ms-xy/Holmes-Planner-Monitor/go/client/diskinfo"
	"io/ioutil"
	"syscall"
)

var (
	SI_LOAD_SHIFT = float64(1 << C.SI_LOAD_SHIFT)
)

// ~ 200 ns/op
// Too bad that inaccuracy of the memory values renders it useless
// for UpdateMeminfo. Load values and uptime appear to be accurate though.
func (this *Sysinfo) UpdateSysinfo() {

	si := &syscall.Sysinfo_t{}

	err := syscall.Sysinfo(si)
	if err != nil {
		this.LastError = err
		return
	}

	this.System.Load[0] = float64(si.Loads[0]) / SI_LOAD_SHIFT
	this.System.Load[1] = float64(si.Loads[1]) / SI_LOAD_SHIFT
	this.System.Load[2] = float64(si.Loads[2]) / SI_LOAD_SHIFT

	// this.System.Uptime = time.Duration(si.Uptime) * time.Second
	this.System.Uptime = si.Uptime
}

// ~ 100 µs/op
// As long as we don't need any other CPU information,
// the little information yield of this one is acceptable.
// Alternative is to just use the goprocs solution used in other functions here.
func (this *Sysinfo) UpdateCores() {
	data, err := ioutil.ReadFile("/proc/cpuinfo")
	if err != nil {
		this.LastError = errors.New("Error reading /proc/cpuinfo: " + err.Error())
		return
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

// ~ 100 µs/op
// Even though the used framework is inefficient, this is still pretty fast.
func (this *Sysinfo) UpdateMeminfo() {
	mi, err := goprocinfo.ReadMemInfo("/proc/meminfo")
	if err != nil {
		this.LastError = err
		return
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
}

// ~ 500 µs/op
// Reason for the somewhat worse performance compared to other functions is
// probably all the reallocations paired with inefficient conversions (not to
// mention the unnecessary regular expressions).
// (The used framework does e.g. append to a list of structs (not struct ptrs))

func (this *Sysinfo) UpdateCpuinfo() {
	ci, err := goprocinfo.ReadCPUInfo("/proc/cpuinfo")
	if err == nil {
		this.Cpu.Cores = ci.NumCPU()
	} else {
		this.LastError = err
	}
}

// ~ 1 ms/op
// Update the harddrive information of the system.
// This function is very slow as it needs to launch a process and read its
// output.
func (this *Sysinfo) UpdateDiskinfo() {
	harddrives, err := diskinfo.Get()
	if err == nil {
		this.Harddrives = harddrives
	} else {
		this.LastError = err
	}
}

// ~ 100 µs/op
// Read /proc/stat to obtain detailed cpu info
// Calculation details taken from:
// http://stackoverflow.com/a/23376195
func (this *Sysinfo) UpdateCpuUsage() {
	stat, err := goprocinfo.ReadStat("/proc/stat")
	if err != nil {
		this.LastError = err
		return
	}

	x := stat.CPUStatAll
	iowait := x.IOWait
	idle := x.Idle + x.IOWait
	busy := x.User + x.Nice + x.System + x.IRQ + x.SoftIRQ + x.Steal
	total := idle + busy

	d_iowait := iowait - this.Cpu.prev_iowait
	d_idle := idle - this.Cpu.prev_idle
	d_busy := busy - this.Cpu.prev_busy
	d_total := total - this.Cpu.prev_total

	// Update struct fields
	// this.Cpu.Load = float64(d_total-d_idle) / float64(d_total) * 100
	this.Cpu.IOWait = d_iowait
	this.Cpu.Idle = d_idle
	this.Cpu.Busy = d_busy
	this.Cpu.Total = d_total

	this.Cpu.prev_iowait = iowait
	this.Cpu.prev_idle = idle
	this.Cpu.prev_busy = busy
	this.Cpu.prev_total = total

	// Alternative real time measurement as root (or sysdig group):
	// launch process "sudo sysdig -c sysinfo_cpu"
	// launch go routine to read output and update self
	// wait for <-this.quit
	// kill process
}
