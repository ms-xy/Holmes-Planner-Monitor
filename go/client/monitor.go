package client

import (
	"github.com/golang/protobuf/proto"
	Netinfo "github.com/ms-xy/Holmes-Planner-Monitor/go/client/netinfo"
	Sysinfo "github.com/ms-xy/Holmes-Planner-Monitor/go/client/sysinfo"
	"github.com/ms-xy/Holmes-Planner-Monitor/go/msgtypes"
	pb "github.com/ms-xy/Holmes-Planner-Monitor/protobuf/generated-go"

	"errors"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"

	"strings"
)

var (
	// Connection, addresses and a buffer for reading, its size set to exactly
	// 65000
	pid        = uint64(os.Getpid())
	uuid       = msgtypes.UUID4Empty()
	raddr      *net.UDPAddr
	laddr      *net.UDPAddr
	connection *net.UDPConn
	buffer     []byte = make([]byte, 0xfde8)
	connected  bool

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
)

// Connect to the remote address using the given local address and try init
// status connection with the given planner name.
// Addresses must have the format host:port or [ipv6-host%zone]:port.
// Terminate an old connection if exists.
//
// TODO: implement UUID transfer, save in /var/cache/Holmes-Processing/machine.uuid
//
func Connect(remoteAddr string, info *msgtypes.PlannerInfo) error {
	// To avoid race conditions we have to check whether or not we are allowed to
	// continue.
	ok, state := stateTransition(StateDisconnected, StateConnecting)
	if !ok {
		logf(LogLevelDebug, "Cannot connect because monitor is %s", state.String())
		return errors.New("Cannot connect because monitor is " + state.String())
	}

	var (
		// localAddr int
		err  error
		ackd bool
	)

	logf(LogLevelDebug, "Initializing sysinfo")
	sysinfo, err = Sysinfo.New()
	if err != nil {
		return err
	}
	logf(LogLevelDebug, "Initializing netinfo")
	netinfo, err = Netinfo.Get()
	if err != nil {
		return err
	}

	logf(LogLevelDebug, "Connecting to %s", remoteAddr)
	if raddr, err = net.ResolveUDPAddr("udp", remoteAddr); err == nil {
		connection, err = net.DialUDP("udp", nil, raddr)
	}
	if err != nil {
		logf(LogLevelDebug, "Connection attempt failed: %s", err.Error())
		stateTransition(StateConnecting, StateDisconnected)
		return err
	}

	// Get the uuid if it already exists
	// TODO: make the location configurable
	var (
		// uuid_file_path string = "/var/cache/Holmes-Processing/uuid"
		uuid_file_path string = "/var/tmp/holmes_processing_cache/uuid"
	)

	// Create the folder in /var/cache
	err = os.MkdirAll(filepath.Dir(uuid_file_path), 0700)
	if err != nil {
		return err
	}

	// Use os.Stat() to determine whether or not the file exists (and is no dir)
	if fi, err := os.Stat(uuid_file_path); err == nil {
		if fi.IsDir() {
			return errors.New(uuid_file_path + ": expected a file, found a directory")
		}
		bytes, err := ioutil.ReadFile(uuid_file_path)
		if err != nil {
			return err
		}
		if len(bytes) > 0 {
			err := uuid.FromString(string(bytes))
			if err != nil {
				return errors.New(uuid_file_path + " contains a malformed uuid: " + err.Error())
			}
			logf(LogLevelDebug, "UUID=%s", uuid.ToString())
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	// Send planner info, expect reply with either matching uuid or new uuid.
	// TODO: configuration option / parameter for number of retries and timeout
	interval := 2 * time.Second
	maxRetry := 10
	logf(LogLevelDebug, "Attempt to connect to Holmes-Status, %d retries, %s timeout", maxRetry, interval.String())

	msg := &pb.StatusMessage{PlannerInfo: info.ToPb()}
	fn := func(resp *msgtypes.ControlMessage) bool {
		if resp.UUID.IsValid() {
			uuid = resp.UUID
		}
		return resp.AckConnect
	}
	ackd, err = sendUntil(msg, fn, interval, maxRetry)

	// If no acknowledge was received, the connection attempt has failed.
	if !ackd {
		if err == nil {
			// TODO: this line definitly needs an improvement
			// could be a connection problem too, not just no ack received
			err = errors.New("Status Server: Ack=False")
		}
		connection.Close()
		stateTransition(StateConnecting, StateDisconnected)
		return err
	}
	logf(LogLevelDebug, "Status Server: Ack=True, UUID=%s", uuid.ToString())

	// Write uuid file
	// TODO: what to do in case of different previous content in this file?
	//       That would be a severe error ...
	err = ioutil.WriteFile(uuid_file_path, []byte(uuid.ToString()), 0600)
	if err != nil {
		return err
	}

	// Create channels and launch incoming as well as outgoing message loops
	// The quit channel is just a cheap throwaway channel to interrupt the
	// loops
	statusMessageChannel = make(chan *pb.StatusMessage, 1000)
	controlMessageChannel = make(chan *msgtypes.ControlMessage, 1000)
	go statusMessageLoop()
	go controlMessageLoop()

	// Start loop to gather some status automatically
	go automaticStatusLoop()

	// data structures to handle disconnecting
	disconnect = make(chan struct{})
	disconnectWaitGroup = &sync.WaitGroup{}

	// signal connection state
	connected = true
	log(LogLevelDebug, "Connected")
	stateTransition(StateConnecting, StateConnected)
	return nil
}

// Disconnect from the status server.
// This should be called upon shutting down the planner, to avoid detection of
// a node-down event on the status server side. (Cannot distinguish a willful
// disconnect from a crash without prior notification)
func Disconnect() error {
	// Advance to disconnecting state only if connected
	ok, state := stateTransition(StateConnected, StateDisconnecting)
	if !ok {
		logf(LogLevelDebug, "Cannot disconnect because monitor is %s", state.String())
		return errors.New("Cannot disconnect because monitor is " + state.String())
	}

	// TODO: Implmement a graceful disconnect from the server
	// TODO: send termination package, requires an ACK to work properly
	// TODO: but only with max retries just like for connect

	// Interupt loops
	disconnectWaitGroup.Add(3) // TODO: how many routines do we actually have?
	connected = false          // this prevents any more messages to be sent
	close(disconnect)
	disconnectWaitGroup.Wait()
	close(statusMessageChannel)
	close(controlMessageChannel)

	// Close the connection and signal connection state
	connection.Close()
	stateTransition(StateDisconnecting, StateDisconnected)
	log(LogLevelDebug, "Disconnected")
	return nil
}

// -----------------------------------------------------------------------------

func send(msg *pb.StatusMessage) error {
	msg.PID = pid
	msg.UUID = uuid.ToBytes()
	bytes, err := proto.Marshal(msg)
	if err == nil {
		_, err = connection.Write(bytes)
	}
	return err
}

type singleMsgStruct struct {
	err error
	msg *msgtypes.ControlMessage
}

func recv(singleMsgChan chan singleMsgStruct) {
	var (
		n    int
		err  error
		rmsg = &msgtypes.ControlMessage{}
	)
	// TODO: what happens if connection is closed to read?
	if n, err = connection.Read(buffer); err == nil {
		if n > 0 {
			msg := &pb.ControlMessage{}
			if err = proto.Unmarshal(buffer[0:n], msg); err == nil {
				rmsg.FromPb(msg)
			}
		}
	}
	singleMsgChan <- singleMsgStruct{err, rmsg}
}

func sendUntil(msg *pb.StatusMessage, fn func(*msgtypes.ControlMessage) bool, interval time.Duration, maxRetry int) (bool, error) {
	singleMsgChan := make(chan singleMsgStruct, 1)
	defer close(singleMsgChan)

	// Remove read deadline ((time.Time{}).IsZero() == true)
	defer connection.SetReadDeadline(time.Time{})

	var err error

	for i := 0; i < maxRetry; i++ {
		if err = send(msg); err == nil {

			connection.SetReadDeadline(time.Now().Add(interval))

			go recv(singleMsgChan)
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

func statusMessageLoop() {
	var err error

	for {

		select {
		case <-disconnect:
			log(LogLevelDebug, "++ Exit (statusMessageLoop)")
			disconnectWaitGroup.Done()
			return

		case msg := <-statusMessageChannel:
			if err = send(msg); err != nil {
				if strings.Contains(err.Error(), "connection refused") {
					log(LogLevelDebug, "-- Connection refused, disconnecting")
					go Disconnect()
				} else {
					logf(LogLevelDebug, "-- Error sending control message: '%v'", err)
				}
			}
		}
	}
}

func controlMessageLoop() {
	singleMsgChan := make(chan singleMsgStruct, 1)
	// defer close(singleMsgChan) // TODO: close it, but needs to be done in recv

	for {
		if !connected {
			log(LogLevelDebug, "++ Exit (controlMessageLoop)")
			disconnectWaitGroup.Done()
			return
		}

		go recv(singleMsgChan)

		select {
		case <-disconnect:
			log(LogLevelDebug, "++ Exit (controlMessageLoop)")
			disconnectWaitGroup.Done()
			return

		case r := <-singleMsgChan:
			if r.err != nil {
				if strings.Contains(r.err.Error(), "connection refused") {
					log(LogLevelDebug, "-- Connection refused, disconnecting")
					go Disconnect()
				} else {
					logf(LogLevelDebug, "-- Error receiving control message: '%v'  %v", r.err, r.msg)
				}
			} else if r.msg != nil {
				// if recv times out, (nil, nil) is returned
				controlMessageChannel <- r.msg
			}
		}

	}
}

// -----------------------------------------------------------------------------

func automaticStatusLoop() {
	systemStatus(&msgtypes.SystemStatus{
		Uptime:      sysinfo.System.Uptime,
		MemoryUsage: sysinfo.Ram.Used,
		MemoryMax:   sysinfo.Ram.Total,
		Harddrives:  sysinfo.Harddrives,
		Loads1:      sysinfo.System.Load[0],
		Loads5:      sysinfo.System.Load[1],
		Loads15:     sysinfo.System.Load[2],
	})
	networkStatus(&msgtypes.NetworkStatus{
		Interfaces: netinfo,
	})
	i := 0
	loopMax := 100
	for {
		select {
		case <-disconnect:
			log(LogLevelDebug, "++ Exit (automaticStatusLoop)")
			disconnectWaitGroup.Done()
			return
		case <-time.After(5 * time.Second):
			// Send an update about the system status
			// Not updating cores as they cannot change at runtime? (TODO: verify!)
			// TODO: treat potential errors returned from both functions
			sysinfo.UpdateMeminfo()
			sysinfo.UpdateSysinfo()
			if i%4 == 0 {
				sysinfo.UpdateDiskinfo()
			}
			systemStatus(&msgtypes.SystemStatus{
				Uptime:      sysinfo.System.Uptime,
				MemoryUsage: sysinfo.Ram.Used,
				MemoryMax:   sysinfo.Ram.Total,
				Harddrives:  sysinfo.Harddrives,
				Loads1:      sysinfo.System.Load[0],
				Loads5:      sysinfo.System.Load[1],
				Loads15:     sysinfo.System.Load[2],
			})
			if i%4 == 0 {
				// TODO: treat potential error
				netinfo, _ = Netinfo.Get()
				networkStatus(&msgtypes.NetworkStatus{
					Interfaces: netinfo,
				})
			}
			if i%100 == 0 {
				debug.FreeOSMemory()
			}
			i = (i + 1) % loopMax
		}
	}
}

// -----------------------------------------------------------------------------

func enqueue(msg *pb.StatusMessage) {
	if !connected {
		log(LogLevelDebug, "++ Deny (enqueue)")
		return
	}
	select {
	case <-disconnect:
		// in case the queue is full when connection is terminated to avoid
		// orphaned goroutines lingering around, as well as panics, return
		log(LogLevelDebug, "++ Exit (enqueue)")
		return
	case statusMessageChannel <- msg:
	}
}

// -----------------------------------------------------------------------------

// The system status and network status types are gathered automatically, as
// such they use the less convenient, but easier maintainable interface.
// Further for that very reason they are not public.
func systemStatus(msg *msgtypes.SystemStatus) {
	enqueue(&pb.StatusMessage{SystemStatus: msg.ToPb()})
}
func networkStatus(msg *msgtypes.NetworkStatus) {
	enqueue(&pb.StatusMessage{NetworkStatus: msg.ToPb()})
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

func PlannerStatus(configProfileName string, logMessages []string, extraData [][]byte) {
	enqueue(&pb.StatusMessage{PlannerStatus: (&msgtypes.PlannerStatus{
		ConfigProfileName: configProfileName,
		Logs:              logMessages,
		ExtraData:         extraData,
	}).ToPb()})
}

func ServiceStatus(configProfileName string, logMessages []string, extraData [][]byte,
	name string, port uint16, task string) {

	enqueue(&pb.StatusMessage{ServiceStatus: (&msgtypes.ServiceStatus{
		ConfigProfileName: configProfileName,
		Name:              name,
		Port:              port,
		Task:              task,
		Logs:              logMessages,
		ExtraData:         extraData,
	}).ToPb()})
}

// -----------------------------------------------------------------------------

func IncomingControlMessages() chan *msgtypes.ControlMessage {
	return controlMessageChannel
}
