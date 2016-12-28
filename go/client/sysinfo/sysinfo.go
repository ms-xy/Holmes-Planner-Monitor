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
	"time"
)

// This struct describes the system information that we want to send to Status
// in order to record it.
// Additionally to the functions below, OS specific functions need to be
// implemented:
//
//   - Sysinfo.UpdateSysinfo():
//		 	System load information
//   - Sysinfo.UpdateCores():
//			Set the number of available cpu cores, this function will at some later
//			point be replaced by a UpdateCpuinfo() function that gathers more
//			information that this one does
//	 - Sysinfo.UpdateMeminfo():
//			Get the amount of memory used (and available), this includes swap usage
//			if applicable to the system
//	 - Sysinfo.UpdateDiskinfo():
//			Update the information about available storage devices
//	 - Sysinfo.UpdateCpuUsage():
//			Get the usage of the CPU in percent, this function is separated from
//			UpdateCores() for the simple reason that we execute it a lot more
//			frequent
//
// If any of these functions encounters an error, it has to set this.LastError,
// This makes error detection a bit easier. Could also add a channel to deal
// with errors, but all that matters is knowing that there is an error (any
// error in this module must be considered fatal, it indicates problems with the
// machine).
//
type Sysinfo struct {
	sync.Mutex

	ticker    *time.Ticker
	quit      chan struct{}
	LastError error

	System struct {
		Uptime int64 //time.Duration
		Load   [3]float64
	}

	Cpu struct {
		IOWait uint64 // this field may indicate why load averages are so high (lots of procs waiting)
		Idle   uint64 // these 3 values are required to calculate the "Load" percentage
		Busy   uint64 // load = (total - idle) / total
		Total  uint64

		prev_iowait uint64
		prev_idle   uint64
		prev_busy   uint64
		prev_total  uint64

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

	// for initialization run updates in parallel and wait for all data to be
	// gathered
	wg := &sync.WaitGroup{}
	wg.Add(4)

	go func() {
		defer wg.Done()
		si.UpdateMeminfo()
	}()

	go func() {
		defer wg.Done()
		si.UpdateDiskinfo()
	}()

	go func() {
		defer wg.Done()
		// si.UpdateCpuinfo()
		si.UpdateCores()
	}()

	go func() {
		defer wg.Done()
		si.UpdateSysinfo()
	}()

	wg.Wait()

	return si, si.LastError
}

func (this *Sysinfo) StartUpdate(every time.Duration) {
	this.Lock()
	defer this.Unlock()
	this.stop()
	go this.update(every)
}

func (this *Sysinfo) StopUpdate() {
	this.Lock()
	defer this.Unlock()
	this.stop()
	this.ticker = nil
}

func (this *Sysinfo) update(every time.Duration) {
	this.ticker = time.NewTicker(every)
	this.quit = make(chan struct{})
	for range this.ticker.C {
		// regular update only includes values that can possibly change:
		// - RAM usage
		// - CPU usage
		// - Disk / Swap usage
		// - Sysinfo (loads etc)
		// the CPU isn't measured by this update function but updates
		// every second based on sysdig input (for Linux)
		go this.UpdateSysinfo()
		go this.UpdateCpuUsage()
		go this.UpdateMeminfo()
		go this.UpdateDiskinfo()
	}
	close(this.quit)
}

func (this *Sysinfo) stop() {
	if this.ticker != nil {
		this.ticker.Stop()
		<-this.quit
	}
}
