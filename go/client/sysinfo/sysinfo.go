// This package extends the idea found at https://github.com/capnm/sysinfo
//
// Prerequisites:
//
// Limitations:
// - currently only implemented for Linux
//
package sysinfo

import (
	"github.com/ms-xy/Holmes-Planner-Monitor/go/msgtypes"
	"sync"
)

type Sysinfo struct {
	System struct {
		Uptime int64 //time.Duration
		Load   [3]float64
	}

	Cpu struct {
		Cores int // (not necessarily phyical cores)
	}

	Ram struct {
		Total     uint64
		Free      uint64
		Available uint64
		Used      uint64
	}

	Swap struct {
		Total     uint64
		Free      uint64
		Available uint64
		Used      uint64
	}

	Harddrives []*msgtypes.Harddrive
}

func New() (*Sysinfo, error) {

	si := &Sysinfo{}

	// for initialization run updates in parallel
	var (
		err_meminfo  error
		err_diskinfo error
		err_cpuinfo  error
		err_sysinfo  error
		wg           = &sync.WaitGroup{}
	)
	wg.Add(4)
	go func() {
		defer wg.Done()
		err_meminfo = si.UpdateMeminfo()
	}()
	go func() {
		defer wg.Done()
		err_diskinfo = si.UpdateDiskinfo()
	}()
	go func() {
		defer wg.Done()
		// err_cpuinfo = si.UpdateCpuinfo()
		err_cpuinfo = si.UpdateCores()
	}()
	go func() {
		defer wg.Done()
		err_sysinfo = si.UpdateSysinfo()
	}()
	wg.Wait()

	if err_meminfo != nil || err_cpuinfo != nil || err_sysinfo != nil {
		var errmsg string = ""
		if err_meminfo != nil {
			errmsg = errmsg + "Error updating meminfo: " + err_meminfo.Error()
		}
		if err_diskinfo != nil {
			errmsg = errmsg + "Error updating diskinfo: " + err_diskinfo.Error()
		}
		if err_cpuinfo != nil {
			errmsg = errmsg + "Error updating cpuinfo: " + err_cpuinfo.Error()
		}
		if err_sysinfo != nil {
			errmsg = errmsg + "Error updating sysinfo: " + err_sysinfo.Error()
		}
	}

	return si, nil
}
