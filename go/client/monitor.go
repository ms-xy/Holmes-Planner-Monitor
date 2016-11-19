package client

import (
	"github.com/golang/protobuf/proto"
	"github.com/ms-xy/Holmes-Planner-Monitor/go/msgtypes"
	pb "github.com/ms-xy/Holmes-Planner-Monitor/protobuf/generated-go"

	"errors"
	"net"
	"time"
)

var (
	// Connection, addresses and a buffer for reading, its size set to exactly
	// 65000
	raddr      *net.UDPAddr
	laddr      *net.UDPAddr
	connection *net.UDPConn
	buffer     []byte = make([]byte, 0xfde8)
	connected  bool

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

	// Names for Messages
	plannerInfo *pb.PlannerInfo
)

// Connect to the remote address using the given local address and try init
// status connection with the given planner name.
// Addresses must have the format host:port or [ipv6-host%zone]:port.
// Terminate an old connection if exists.
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
		cmsg *msgtypes.ControlMessage
	)
	// Connect to the server
	logf(LogLevelDebug, "Connecting to %s", remoteAddr)
	if raddr, err = net.ResolveUDPAddr("udp", remoteAddr); err == nil {
		connection, err = net.DialUDP("udp", nil, raddr)
	}
	if err != nil {
		logf(LogLevelDebug, "Connection attempt failed: %s", err.Error())
		stateTransition(StateConnecting, StateDisconnected)
		return err
	}
	// Send PlannerInfo and await acknowledge
	// At max 10 retries
	// TODO: configuration option / parameter for number of retries and timeout
	plannerInfo = info.ToPb()
	// Send the PlannerInfo every 2 seconds until it is ack'd for a max of 10
	// retries
	ackd = false
	for i := 0; i < 10; i++ {
		if err = send(&pb.StatusMessage{PlannerInfo: plannerInfo}); err == nil {
			logf(LogLevelDebug, "Sending planner information (attempt %d)", i+1)
			connection.SetReadDeadline(time.Now().Add(time.Duration(2 * time.Second)))
			if cmsg, err = recv(); err == nil {
				if cmsg.AckConnect {
					ackd = true
				} else {
					log(LogLevelDebug, "Status module responded with ACK != true to planner information")
				}
				break
			} else {
				log(LogLevelDebug, "Failed to receive ACK-response: "+err.Error())
			}
		} else {
			log(LogLevelDebug, "Failed to send planner information: "+err.Error())
		}
	}
	// Remove read deadline ((time.Time{}).IsZero() == true)
	connection.SetReadDeadline(time.Time{})
	// Check Ack status, if Ack failed, connection is assumed failed too
	if !ackd {
		if err == nil {
			// TODO: this line definitly needs an upgrade, can also be a connection
			// failure
			err = errors.New("Status Server: Ack=False")
		}
		connection.Close()
		connection = nil
		stateTransition(StateConnecting, StateDisconnected)
		return err
	}
	log(LogLevelDebug, "Received ACK for planner information")
	// Create channels and launch incoming as well as outgoing message loops
	statusMessageChannel = make(chan *pb.StatusMessage, 1000)
	controlMessageChannel = make(chan *msgtypes.ControlMessage, 1000)
	go statusMessageLoop()
	go controlMessageLoop()
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
	// Close all channels, this stops the loops as well
	close(statusMessageChannel)
	close(controlMessageChannel)
	connected = false
	connection.Close()
	connection = nil
	stateTransition(StateDisconnecting, StateDisconnected)
	return nil
}

// -----------------------------------------------------------------------------

func send(msg *pb.StatusMessage) error {
	bytes, err := proto.Marshal(msg)
	if err == nil {
		_, err = connection.Write(bytes)
	}
	return err
}

func recv() (*msgtypes.ControlMessage, error) {
	var (
		err  error
		rmsg = &msgtypes.ControlMessage{}
	)
	if n, err := connection.Read(buffer); err == nil {
		if n > 0 {
			msg := &pb.ControlMessage{}
			if err = proto.Unmarshal(buffer[0:n], msg); err == nil {
				rmsg.FromPb(msg)
			}
		}
	}
	return rmsg, err
}

// -----------------------------------------------------------------------------

func statusMessageLoop() {
	for {
		msg, ok := <-statusMessageChannel
		if !ok {
			return
		}
		send(msg)
		// TODO: treat possible error return from send
	}
}

func controlMessageLoop() {
	for {
		msg, _ := recv()
		if msg != nil { // if recv times out, (nil, nil) is returned
			// TODO: what happens if connection is closed to read? and subsequently,
			// will this assignment fail? (probably panic?)
			controlMessageChannel <- msg
		}
		// TODO: treat possible error return from recv
	}
}

// -----------------------------------------------------------------------------

func SystemStatus(msg *msgtypes.SystemStatus) {
	statusMessageChannel <- &pb.StatusMessage{SystemStatus: msg.ToPb()}
}

func NetworkStatus(msg *msgtypes.NetworkStatus) {
	statusMessageChannel <- &pb.StatusMessage{NetworkStatus: msg.ToPb()}
}

func PlannerStatus(msg *msgtypes.PlannerStatus) {
	statusMessageChannel <- &pb.StatusMessage{PlannerStatus: msg.ToPb()}
}

func ServiceStatus(msg *msgtypes.ServiceStatus) {
	statusMessageChannel <- &pb.StatusMessage{ServiceStatus: msg.ToPb()}
}

func IncomingControlMessages() chan *msgtypes.ControlMessage {
	return controlMessageChannel
}
