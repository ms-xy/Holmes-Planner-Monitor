package client

import (
	"github.com/golang/protobuf/proto"
	Netinfo "github.com/ms-xy/Holmes-Planner-Monitor/go/client/netinfo"
	Sysinfo "github.com/ms-xy/Holmes-Planner-Monitor/go/client/sysinfo"
	"github.com/ms-xy/Holmes-Planner-Monitor/go/msgtypes"
	pb "github.com/ms-xy/Holmes-Planner-Monitor/protobuf/generated-go"

	"errors"
	"log"
	"net"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"strings"
)

var (
	singleton *Monitor
)

func GetInstance() *Monitor {
	if singleton == nil {
		singleton = NewInstance()
	}
	return singleton
}

func NewInstance() *Monitor {
	m := &Monitor{}
	m.Init()
	return m
}

type Monitor struct {
	// Inherit from state manager and logger
	MonitorStateManager
	Logger
	// Connection, addresses and a buffer for reading, its size set to exactly
	// 65000
	pid            uint64
	uuid           *msgtypes.UUID
	machine_uuid   *msgtypes.UUID
	uuid_file_path string
	raddr          *net.UDPAddr
	laddr          *net.UDPAddr
	connection     *net.UDPConn
	buffer         []byte
	connected      bool

	// Automatic status information gathering (sysinfo, meminfo, cpuinfo)
	sysinfo *Sysinfo.Sysinfo
	netinfo []*msgtypes.NetworkInterface

	// Data Transfer
	//
	// Currently supported message types are:
	//     - SystemStatus
	//     - NetworkStatus
	//     - PlannerStatus
	//     - ServiceStatus
	//
	// See /protobuf/messages.proto for details.
	// All fields are optional and empty fields are interpreted as not set.
	// Incoming control messages are somewhat special. Each may be dispatched
	// to a callback function multiple times. Callbacks are not threadsafe or
	// anything, locks / semaphores need to applied as appropriate.
	//
	statusMessageChannel  chan *pb.StatusMessage
	controlMessageChannel chan *msgtypes.ControlMessage
	disconnect            chan struct{}
	disconnectWaitGroup   *sync.WaitGroup

	// Names for Messages
	plannerInfo *pb.PlannerInfo
}

func (this *Monitor) Init() {
	// Init logger (with prefix)
	// Default log level is quiet, equivalent to no logging
	this.LogOutput = log.New(os.Stdout, "Status-Monitor: ", log.Ldate|log.Ltime|log.Lshortfile)
	this.LogLevel = LogLevelQuiet

	// Init monitor
	this.pid = uint64(os.Getpid())
	this.uuid = msgtypes.UUID4Empty()
	this.machine_uuid = msgtypes.UUID4Empty()
	this.buffer = make([]byte, 0xfde8)
	this.connected = false
	// TODO: make the location configurable
	// this.uuid_file_path = "/var/cache/Holmes-Processing/uuid"
	this.uuid_file_path = "/var/tmp/holmes_processing_cache/uuid"

	// Init state manager
	this.monitorlock = &sync.Mutex{}
	this.monitorstate = StateDisconnected
}

// Connect to the remote address using the given local address and try init
// status connection with the given planner name.
// Addresses must have the format host:port or [ipv6-host%zone]:port.
// Terminate an old connection if exists.
//
func (this *Monitor) Connect(remoteAddr string, info *msgtypes.PlannerInfo) error {
	// To avoid race conditions we have to check whether or not we are allowed to
	// continue.
	ok, state := this.stateTransition(StateDisconnected, StateConnecting)
	if !ok {
		this.Logf(LogLevelDebug, "Cannot connect because monitor is %s", state.String())
		return errors.New("Cannot connect because monitor is " + state.String())
	}

	var (
		// localAddr int
		err  error
		ackd bool
	)

	this.Logf(LogLevelDebug, "Initializing sysinfo")
	this.sysinfo, err = Sysinfo.New()
	this.sysinfo.StartUpdate(1 * time.Second)
	if err != nil {
		return err
	}
	this.Logf(LogLevelDebug, "Initializing netinfo")
	this.netinfo, err = Netinfo.Get()
	if err != nil {
		return err
	}

	this.Logf(LogLevelDebug, "Connecting to %s", remoteAddr)
	if this.raddr, err = net.ResolveUDPAddr("udp", remoteAddr); err == nil {
		this.connection, err = net.DialUDP("udp", nil, this.raddr)
	}
	if err != nil {
		this.Logf(LogLevelDebug, "Connection attempt failed: %s", err.Error())
		this.stateTransition(StateConnecting, StateDisconnected)
		return err
	}

	// Load machine uuid
	if err := this.load_machine_uuid(); err != nil {
		return err
	}

	// Send planner info, expect reply with either matching uuid or new uuid.
	// TODO: configuration option / parameter for number of retries and timeout
	interval := 10 * time.Second
	maxRetry := 10
	this.Logf(LogLevelDebug, "Attempt to connect to Holmes-Status, %d retries, %s timeout",
		maxRetry,
		interval.String())

	msg := &pb.StatusMessage{PlannerInfo: info.ToPb()}
	fn := func(resp *msgtypes.ControlMessage) bool {
		this.Logf(LogLevelDebug, "Received a control-message: %v", resp)
		if resp.UUID.IsValid() {
			this.uuid = resp.UUID
		}
		if resp.MachineUUID.IsValid() {
			this.machine_uuid = resp.MachineUUID
		}
		return resp.AckConnect
	}
	ackd, err = this.sendUntil(msg, fn, interval, maxRetry)

	// If no acknowledge was received, the connection attempt has failed.
	if !ackd {
		if err == nil {
			// TODO: this line definitly needs an improvement
			// could be a connection problem too, not just no ack received
			err = errors.New("Status Server: Ack=False")
		}
		this.connection.Close()
		this.stateTransition(StateConnecting, StateDisconnected)
		return err
	}
	this.Logf(LogLevelDebug, "Status Server: Ack=True, MachineUUID=%s, UUID=%s",
		this.machine_uuid.ToString(),
		this.uuid.ToString())

	// Save machine uuid
	err = this.save_machine_uuid()

	// Create channels and launch incoming as well as outgoing message loops
	// The quit channel is just a cheap throwaway channel to interrupt the
	// loops
	this.statusMessageChannel = make(chan *pb.StatusMessage, 1000)
	this.controlMessageChannel = make(chan *msgtypes.ControlMessage, 1000)
	go this.statusMessageLoop()
	go this.controlMessageLoop()

	// Start loop to gather some status automatically
	go this.automaticStatusLoop()

	// data structures to handle disconnecting
	this.disconnect = make(chan struct{})
	this.disconnectWaitGroup = &sync.WaitGroup{}

	// signal connection state
	this.connected = true
	this.Log(LogLevelDebug, "Connected")
	this.stateTransition(StateConnecting, StateConnected)
	return nil
}

// Disconnect from the status server.
// This should be called upon shutting down the planner, to avoid detection of
// a node-down event on the status server side. (Cannot distinguish a willful
// disconnect from a crash without prior notification)
func (this *Monitor) Disconnect() error {
	// Advance to disconnecting state only if connected
	ok, state := this.stateTransition(StateConnected, StateDisconnecting)
	if !ok {
		this.Logf(LogLevelDebug, "Cannot disconnect because monitor is %s", state.String())
		return errors.New("Cannot disconnect because monitor is " + state.String())
	}

	// TODO: Implmement a graceful disconnect from the server
	// TODO: send termination package, requires an ACK to work properly
	// TODO: but only with max retries just like for connect

	// Interupt loops
	this.disconnectWaitGroup.Add(3) // TODO: how many routines do we actually have?
	this.connected = false          // this prevents any more messages to be sent
	close(this.disconnect)
	this.disconnectWaitGroup.Wait()
	close(this.statusMessageChannel)
	close(this.controlMessageChannel)

	// Close the connection and signal connection state
	this.connection.Close()
	this.stateTransition(StateDisconnecting, StateDisconnected)
	this.Log(LogLevelDebug, "Disconnected")
	return nil
}

// -----------------------------------------------------------------------------

func (this *Monitor) send(msg *pb.StatusMessage) error {
	// Fill mandatory fields. PID, UUID, and MachineUUID allow for specific
	// identification, whilst Timestamp pinpoints events to a specific time,
	// allowing for fine grained statistics and searching.
	// The delay between enqueuing a message and actually sending it should be
	// neglectable in this context.
	// msg.Pid = pid
	msg.Uuid = this.uuid.ToBytes()
	msg.MachineUuid = this.machine_uuid.ToBytes()
	msg.Timestamp = uint64(time.Now().UnixNano())

	bytes, err := proto.Marshal(msg)
	if err == nil {
		// this.Log(LogLevelDebug, "Sending message")
		// debug.PrintStack()
		_, err = this.connection.Write(bytes)
	}
	return err
}

type singleMsgStruct struct {
	err error
	msg *msgtypes.ControlMessage
}

func (this *Monitor) recv(singleMsgChan chan singleMsgStruct) {
	var (
		n    int
		err  error
		rmsg = &msgtypes.ControlMessage{}
	)
	// TODO: what happens if connection is closed to read?
	if n, err = this.connection.Read(this.buffer); err == nil {
		if n > 0 {
			msg := &pb.ControlMessage{}
			if err = proto.Unmarshal(this.buffer[0:n], msg); err == nil {
				rmsg.FromPb(msg)
			}
		}
	}
	singleMsgChan <- singleMsgStruct{err, rmsg}
}

func (this *Monitor) sendUntil(msg *pb.StatusMessage, fn func(*msgtypes.ControlMessage) bool, interval time.Duration, maxRetry int) (bool, error) {
	singleMsgChan := make(chan singleMsgStruct, 1)
	defer close(singleMsgChan)

	// Remove read deadline ((time.Time{}).IsZero() == true)
	defer this.connection.SetReadDeadline(time.Time{})

	var err error

	for i := 0; i < maxRetry; i++ {

		// this.Log(LogLevelDebug, "attempt sending")

		if err = this.send(msg); err == nil {

			this.connection.SetReadDeadline(time.Now().Add(interval))
			go this.recv(singleMsgChan)
			r := <-singleMsgChan

			if r.err == nil {
				if fn(r.msg) {
					return true, nil
				}
			} else {
				err = r.err
			}
		}
	}

	return false, err
}

// -----------------------------------------------------------------------------

func (this *Monitor) statusMessageLoop() {
	var err error

	for {

		select {
		case <-this.disconnect:
			this.Log(LogLevelDebug, "++ Exit (statusMessageLoop)")
			this.disconnectWaitGroup.Done()
			return

		case msg := <-this.statusMessageChannel:
			if err = this.send(msg); err != nil {
				if strings.Contains(err.Error(), "connection refused") {
					this.Log(LogLevelDebug, "-- Connection refused, disconnecting")
					go this.Disconnect()
				} else {
					this.Logf(LogLevelErrors, "-- Error sending status message: '%v'", err)
				}
			}
		}
	}
}

func (this *Monitor) controlMessageLoop() {
	singleMsgChan := make(chan singleMsgStruct, 1)
	// defer close(singleMsgChan) // TODO: close it, but needs to be done in recv

	for {
		if !this.connected {
			this.Log(LogLevelDebug, "++ Exit (controlMessageLoop)")
			this.disconnectWaitGroup.Done()
			return
		}

		go this.recv(singleMsgChan)

		select {
		case <-this.disconnect:
			this.Log(LogLevelDebug, "++ Exit (controlMessageLoop)")
			this.disconnectWaitGroup.Done()
			return

		case r := <-singleMsgChan:
			if r.err != nil {
				if strings.Contains(r.err.Error(), "connection refused") {
					this.Log(LogLevelDebug, "-- Connection refused, disconnecting")
					go this.Disconnect()
				} else {
					this.Logf(LogLevelDebug, "-- Error receiving control message: '%v'  %v", r.err, r.msg)
				}
			} else if r.msg != nil {
				// if recv times out, (nil, nil) is returned
				this.controlMessageChannel <- r.msg
			}
		}

	}
}

// -----------------------------------------------------------------------------

func (this *Monitor) automaticStatusLoop() {
	_send := func(sysinfo *Sysinfo.Sysinfo, netinfo []*msgtypes.NetworkInterface) {
		this.systemStatus(&msgtypes.SystemStatus{
			Uptime: sysinfo.System.Uptime,

			CpuIOWait: sysinfo.Cpu.IOWait,
			CpuIdle:   sysinfo.Cpu.Idle,
			CpuBusy:   sysinfo.Cpu.Busy,
			CpuTotal:  sysinfo.Cpu.Total,

			MemoryUsage: sysinfo.Ram.Used,
			MemoryMax:   sysinfo.Ram.Total,
			SwapUsage:   sysinfo.Swap.Used,
			SwapMax:     sysinfo.Swap.Total,

			Harddrives: sysinfo.Harddrives,

			Loads1:  sysinfo.System.Load[0],
			Loads5:  sysinfo.System.Load[1],
			Loads15: sysinfo.System.Load[2],
		})
		this.networkStatus(&msgtypes.NetworkStatus{
			Interfaces: netinfo,
		})
	}
	// send initial system and network status messages
	_send(this.sysinfo, this.netinfo)
	// regular updates (every second)
	i := 0
	loopMax := 100
	for {
		select {
		case <-this.disconnect:
			this.Log(LogLevelDebug, "++ Exit (automaticStatusLoop)")
			this.disconnectWaitGroup.Done()
			return
		case <-time.After(1 * time.Second):
			// TODO: treat potential error by NetInfo.Get()
			this.netinfo, _ = Netinfo.Get()
			// publish update
			_send(this.sysinfo, this.netinfo)
			// only execute the expensive global free every "loopMax" cycles
			// basically results in a good memory balance and overall cheap cycles
			if i%loopMax == 0 {
				debug.FreeOSMemory()
			}
			i = (i + 1) % loopMax
		}
	}
}

// -----------------------------------------------------------------------------

func (this *Monitor) enqueue(msg *pb.StatusMessage) {
	if !this.connected {
		this.Log(LogLevelDebug, "++ Deny (enqueue)")
		return
	}
	select {
	case <-this.disconnect:
		// in case the queue is full when connection is terminated to avoid
		// orphaned goroutines lingering around, as well as panics, return
		this.Log(LogLevelDebug, "++ Exit (enqueue)")
		return
	case this.statusMessageChannel <- msg:
	}
}

// -----------------------------------------------------------------------------

// The system status and network status types are gathered automatically, as
// such they use the less convenient, but easier maintainable interface.
// Further for that very reason they are not public.
func (this *Monitor) systemStatus(msg *msgtypes.SystemStatus) {
	this.enqueue(&pb.StatusMessage{SystemStatus: msg.ToPb()})
}
func (this *Monitor) networkStatus(msg *msgtypes.NetworkStatus) {
	this.enqueue(&pb.StatusMessage{NetworkStatus: msg.ToPb()})
}

// Two alternatives for planner status, better maintainable version and the
// multi-param version. The latter is easier to use.

// func PlannerStatus(msg *msgtypes.PlannerStatus) {
// 	enqueue(&pb.StatusMessage{PlannerStatus: msg.ToPb()})
// }

// Same goes for the service status ...

// func ServiceStatus(msg *msgtypes.ServiceStatus) {
// 	enqueue(&pb.StatusMessage{ServiceStatus: msg.ToPb()})
// }

func (this *Monitor) PlannerStatus(configProfileName string, logMessages []string, extraData [][]byte) {
	this.enqueue(&pb.StatusMessage{PlannerStatus: (&msgtypes.PlannerStatus{
		ConfigProfileName: configProfileName,
		Logs:              logMessages,
		ExtraData:         extraData,
	}).ToPb()})
}

func (this *Monitor) ServiceStatus(configProfileName string, logMessages []string, extraData [][]byte,
	name string, uri string, task string) {

	this.enqueue(&pb.StatusMessage{ServiceStatus: (&msgtypes.ServiceStatus{
		ConfigProfileName: configProfileName,
		Name:              name,
		Uri:               uri,
		Task:              task,
		Logs:              logMessages,
		ExtraData:         extraData,
	}).ToPb()})
}

// -----------------------------------------------------------------------------

func (this *Monitor) IncomingControlMessages() chan *msgtypes.ControlMessage {
	return this.controlMessageChannel
}
