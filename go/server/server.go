package server

import (
	"github.com/golang/protobuf/proto"
	types "github.com/ms-xy/Holmes-Planner-Monitor/go/msgtypes"
	pb "github.com/ms-xy/Holmes-Planner-Monitor/protobuf/generated-go"

	"errors"
	"net"
	// "sync/atomic" // atomic.AddUint64(&last_client_id, 1)
	"time"
)

var (
	laddr      *net.UDPAddr
	connection *net.UDPConn
	inQueue    chan AddressedStatusMessage
	outQueue   chan AddressedControlMessage
	router     StatusRouter
	sessionmap *SessionMap
)

func GetSessions() *SessionMap {
	return sessionmap
}

func ListenAndServe(httpbinding string, statusrouter StatusRouter) error {
	if connection != nil {
		return errors.New("")
	}
	// Start UDP server
	var err error
	if laddr, err = net.ResolveUDPAddr("udp", httpbinding); err == nil {
		connection, err = net.ListenUDP("udp", laddr)
	} else {
		panic(err)
	}
	// initialize variables
	router = statusrouter
	sessionmap = NewSessionMap()
	inQueue = make(chan AddressedStatusMessage, 0x800)
	outQueue = make(chan AddressedControlMessage, 0x800)
	// launch main loops
	go dispatcher()
	go receiver()
	go sender()
	return nil
}

type StatusRouter interface {
	RecvPlannerInfo(*types.PlannerInfo, *Session) *types.ControlMessage
	RecvSystemStatus(*types.SystemStatus, *Session) *types.ControlMessage
	RecvNetworkStatus(*types.NetworkStatus, *Session) *types.ControlMessage
	RecvPlannerStatus(*types.PlannerStatus, *Session) *types.ControlMessage
	RecvServiceStatus(*types.ServiceStatus, *Session) *types.ControlMessage
	HandleError(err error, client *Session) *types.ControlMessage
}

type AddressedStatusMessage struct {
	Address *net.UDPAddr
	Message *types.StatusMessage
}

type AddressedControlMessage struct {
	Address *net.UDPAddr
	Message *types.ControlMessage
}

// Dispatcher that processes incoming messages and queues responses
func dispatcher() {
	var (
		asm AddressedStatusMessage
	)
	for {
		asm = <-inQueue
		// dispatch depending on contained messages
		go func(asm AddressedStatusMessage) {
			now := time.Now()

			// get session instance / create if not exists
			session, isnew := sessionmap.StartSession(asm)

			session.LastSeen = now
			if isnew {
				session.FirstSeen = now
			}

			var (
				cmsg *types.ControlMessage
			)

			if asm.Message.ServiceStatus != nil {
				session.Last.ServiceStatus = now
				cmsg = router.RecvServiceStatus(asm.Message.ServiceStatus, session)

			} else if asm.Message.PlannerStatus != nil {
				session.Last.PlannerStatus = now
				cmsg = router.RecvPlannerStatus(asm.Message.PlannerStatus, session)

			} else if asm.Message.SystemStatus != nil {
				session.Last.SystemStatus = now
				cmsg = router.RecvSystemStatus(asm.Message.SystemStatus, session)

			} else if asm.Message.NetworkStatus != nil {
				session.Last.NetworkStatus = now
				cmsg = router.RecvNetworkStatus(asm.Message.NetworkStatus, session)

			} else if asm.Message.PlannerInfo != nil {
				cmsg = router.RecvPlannerInfo(asm.Message.PlannerInfo, session)
			}

			if cmsg != nil {
				cmsg.UUID = session.GetUuid()
				cmsg.MachineUUID = session.GetMachineUuid()
				outQueue <- AddressedControlMessage{asm.Address, cmsg}
			}
		}(asm)
	}
}

// Blocking reader for the connection.
// Packages are expected to be 65000 bytes or less.
// Does not apply any transformations, that's the job of the dispatcher.
func receiver() {
	var (
		n         int
		addr      *net.UDPAddr
		err       error
		pkgbuffer []byte = make([]byte, 0xfde8) // buffer of size 65000
		buffer    []byte
		msg       *pb.StatusMessage
		asm       AddressedStatusMessage
	)
	for {
		n, addr, err = connection.ReadFromUDP(pkgbuffer)
		if err != nil {
			// TODO error log == read error (push to router too)
			continue
		}
		buffer = make([]byte, n)
		copy(buffer, pkgbuffer[0:n])
		msg = &pb.StatusMessage{}
		err = proto.Unmarshal(buffer, msg)
		if err != nil {
			// TODO error log == invalid message (push to router too)
			continue
		}
		asm = AddressedStatusMessage{
			Message: (&types.StatusMessage{}).FromPb(msg),
			Address: addr,
		}
		inQueue <- asm
	}
}

// Blocking sender for the connection.
// Performs the transformation types.ControlMessage -> pb.ControlMessage.
func sender() {
	var (
		acm   AddressedControlMessage
		bytes []byte
		err   error
	)
	for {
		acm = <-outQueue
		bytes, err = proto.Marshal(acm.Message.ToPb())
		if err != nil {
			// oops? log? (push to router too)
		}
		_, err = connection.WriteToUDP(bytes, acm.Address)
		if err != nil {
			// TODO error log == write error (push to router too)
		}
	}
}
