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
	RecvPlannerInfo(plannerinfo *types.PlannerInfo, client *Session) *types.ControlMessage
	RecvSystemStatus(systemstatus *types.SystemStatus, client *Session) *types.ControlMessage
	RecvNetworkStatus(networkstatus *types.NetworkStatus, client *Session) *types.ControlMessage
	RecvPlannerStatus(plannerstatus *types.PlannerStatus, client *Session) *types.ControlMessage
	RecvServiceStatus(servicestatus *types.ServiceStatus, client *Session) *types.ControlMessage
	HandleError(err error, client *Session)
}

type AddressedStatusMessage struct {
	Address *net.UDPAddr
	Message *pb.StatusMessage
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
			// get session instance / create if not exists
			session := sessionmap.StartSession(asm)
			now := time.Now()

			var (
				cmsg *types.ControlMessage
			)

			session.LastSeen = now

			if asm.Message.ServiceStatus != nil {
				session.Last.ServiceStatus = now
				cmsg = router.RecvServiceStatus((&types.ServiceStatus{}).FromPb(asm.Message.ServiceStatus), session)

			} else if asm.Message.PlannerStatus != nil {
				session.Last.PlannerStatus = now
				cmsg = router.RecvPlannerStatus((&types.PlannerStatus{}).FromPb(asm.Message.PlannerStatus), session)

			} else if asm.Message.SystemStatus != nil {
				session.Last.SystemStatus = now
				cmsg = router.RecvSystemStatus((&types.SystemStatus{}).FromPb(asm.Message.SystemStatus), session)

			} else if asm.Message.NetworkStatus != nil {
				session.Last.NetworkStatus = now
				cmsg = router.RecvNetworkStatus((&types.NetworkStatus{}).FromPb(asm.Message.NetworkStatus), session)

			} else if asm.Message.PlannerInfo != nil {
				session.FirstSeen = now
				cmsg = router.RecvPlannerInfo((&types.PlannerInfo{}).FromPb(asm.Message.PlannerInfo), session)

			}

			if cmsg != nil {
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
		asm       AddressedStatusMessage
	)
	for {
		n, addr, err = connection.ReadFromUDP(pkgbuffer)
		// init asm up here already as error handling requires the sender address
		asm = AddressedStatusMessage{Message: &pb.StatusMessage{}}
		asm.Address = addr
		if err != nil {
			// TODO error log == read error (push to router too)
			continue
		}
		buffer = make([]byte, n)
		copy(buffer, pkgbuffer[0:n])
		err = proto.Unmarshal(buffer, asm.Message)
		if err != nil {
			// TODO error log == invalid message (push to router too)
			continue
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
