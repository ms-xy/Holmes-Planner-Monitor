package server

import (
	"crypto/sha256"
	"encoding/binary"
	"github.com/ms-xy/Holmes-Planner-Monitor/go/msgtypes"
	"net"
	"sync"
)

//
// Session map object
// Contains useful functions for operations on the client cache
//
// map_uuid2machine is the map of all registered unique machines.
// map_uuid2planner on the other hand is the map of all registered planners.
//
func NewSessionMap() *SessionMap {
	return &SessionMap{
		map_uuid2machine: make(map[msgtypes.UUID]map[uint64]*Session),
		map_uuid2planner: make(map[msgtypes.UUID]*Session),
	}
}

type SessionMap struct {
	sync.Mutex
	map_uuid2machine map[msgtypes.UUID]map[uint64]*Session
	map_uuid2planner map[msgtypes.UUID]*Session
}

// Check whether a session exists for the client. If not create a new one.
// Returns the session object as well as a boolean indicating if the client is
// new.
// Sessions are unique per planner but identify with one single machine, given
// by its UUID.
// When a planner connects, if it does not have a machine UUID, one is supplied.
// It should be saved persistently, in a manner such that any other planners
// will be able to use the exact same UUID to be identified correctly.
// Additionally it is given a planner UUID, that it may or may not chose to
// save persistently.
func (this *SessionMap) StartSession(asm AddressedStatusMessage) (*Session, bool) {
	this.Lock()
	defer this.Unlock()

	// Because we abandoned the multiple key approach, we can now simply check if
	// we already know the uuid or not, previously we had to rely on the IP as
	// well as a second identifier.
	var (
		uuid         = asm.Message.UUID
		machine_uuid = asm.Message.MachineUUID
		pid          = asm.Message.PID
	)

	// If the supplied UUIDs are not valid, create new ones.
	// Multipurpose.
	// First, this ensures we never have invalid UUIDs after this point.
	// Second, we can now just use the uuid/machine_uuid as newly created if it
	// does not exist in the respective map.
	if !uuid.IsValid() {
		uuid = this.newUUID()
	}
	if !machine_uuid.IsValid() {
		machine_uuid = this.newUUID()
	}

	planners, exists_machine := this.map_uuid2machine[*machine_uuid]
	session, exists_session := this.map_uuid2planner[*machine_uuid]

	if !exists_machine || !exists_session {
		// Probably a new planner.
		if !exists_session {
			session = &Session{
				machine_uuid: machine_uuid,
				uuid:         uuid,
			}
		}

		// Probably a new machine.
		if !exists_machine {
			// Create a new planners map and assign it.
			planners = make(map[uint64]*Session)
			this.map_uuid2machine[*machine_uuid] = planners

			// Sessions that have been registered on a different machine are moved
			// directly to the new machine. The reasoning behind this is very simple.
			// For example if we have a malfunctioning planner that occupies an
			// UUID that belongs to a different planner, then no data is "really" lost,
			// but rather recorded with alternating machine UUIDs.
			session.machine_uuid = machine_uuid
		}
	}

	// Update maps.
	planners[pid] = session
	this.map_uuid2planner[*uuid] = session

	// Update connection related fields of the session object.
	session.Address = asm.Address
	session.pid = pid

	// Session is only new if it did not exist before (even if it was assigned to
	// a different machine previously)
	return session, !exists_session
}

// Remove a session from the session storage
// This deletes references from all respective maps (machine->planners, and the
// global session map), additionally if the machine->planners map is empty, it
// is removed as well.
func (this *SessionMap) Remove(session *Session) {
	this.Lock()
	this.Unlock()

	session, exist_session := this.map_uuid2planner[*session.uuid]
	if exist_session {
		delete(this.map_uuid2planner, *session.uuid)

		planners, _ := this.map_uuid2machine[*session.machine_uuid]
		delete(planners, session.pid)

		if len(planners) == 0 {
			delete(this.map_uuid2machine, *session.machine_uuid)
		}
	}
}

// Get a session by its uuid
func (this *SessionMap) GetByUuid(uuid *msgtypes.UUID) (*Session, bool) {
	session, exists := this.map_uuid2planner[*uuid]
	return session, exists
}

// Get all planners associated with a machine
func (this *SessionMap) GetAllByMachine(machine_uuid *msgtypes.UUID) (map[uint64]*Session, bool) {
	planners, exist := this.map_uuid2machine[*machine_uuid]
	return planners, exist
}

// Loop over all sessions in the session storage and execute the function fn for
// every single one
func (this *SessionMap) ForEachSession(fn func(*Session)) {
	this.Lock()
	defer this.Unlock()
	for _, session := range this.map_uuid2planner {
		fn(session)
	}
}

// Loop over all machines and execute the function fn for each of them
func (this *SessionMap) ForEachMachine(fn func(*msgtypes.UUID, map[uint64]*Session)) {
	this.Lock()
	defer this.Unlock()
	for machine_uuid, planners := range this.map_uuid2machine {
		fn(&machine_uuid, planners)
	}
}

// Get the size of the session object map
func (this *SessionMap) SizeSessions() int {
	return len(this.map_uuid2planner)
}

// Get the amount of machines registered
func (this *SessionMap) SizeMachines() int {
	return len(this.map_uuid2machine)
}

// -----------------------------------------------------------------------------
// Helper function to create new unique and unused UUIDs
// Note: This function does not lock appropriately, this is the duty of the
//       caller.
// TODO: Better solution in case of an error than to continue and try again?
//       The problem is that if there is an error it is a read error. (reading
//       of /dev/rand failed ... which most likely would be an endless loop?)
func (this *SessionMap) newUUID() *msgtypes.UUID {
	for {
		if uuid, err := msgtypes.UUID4(); err == nil {
			_, exists_machine := this.map_uuid2machine[*uuid]
			_, exists_planner := this.map_uuid2planner[*uuid]
			if !exists_machine && !exists_planner {
				return uuid
			}
		}
	}
}

// -----------------------------------------------------------------------------
// Helper functions to convert input data into values usable as map keys
func addr2bytemap18key(ip net.IP, port int) [18]byte {
	var r [18]byte
	copy(r[:16], (ip.To16())[:16])
	binary.LittleEndian.PutUint16(r[16:18], uint16(port))
	return r
}

func addr2uint64map3key(ip net.IP, port int) [3]uint64 {
	var r [3]uint64
	_ip := ip.To16()
	r[0] = binary.LittleEndian.Uint64(_ip[:8])
	r[1] = binary.LittleEndian.Uint64(_ip[8:])
	r[2] = uint64(port)
	return r
}

func str2bytemap32key(str string) [32]byte {
	return sha256.Sum256([]byte(str))
}

func str2uint64map4key(str string) [4]uint64 {
	bytes := sha256.Sum256([]byte(str))
	key := [4]uint64{}
	for i, p := 0, 0; i < 4; i, p = i+1, p+8 {
		key[i] = binary.LittleEndian.Uint64(bytes[p : p+8])
	}
	return key
}

// -----------------------------------------------------------------------------
// Helper function for ip comparison for the map keys
func bytemap18key_equal(a, b [18]byte) bool {
	for i := 0; i < 18; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func bytemap18key_IP_equal(a, b [18]byte) bool {
	for i := 0; i < 16; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func extractListenAddr(asm AddressedStatusMessage) (listenAddr *net.TCPAddr, exists bool) {
	if asm.Message.PlannerInfo != nil && asm.Message.PlannerInfo.ListenAddress != nil {
		listenAddr = asm.Message.PlannerInfo.ListenAddress
		exists = listenAddr != nil
	}
	return
}

func extractUuid(asm AddressedStatusMessage) (uuid *msgtypes.UUID, exists bool) {
	if asm.Message.PlannerInfo != nil {
		uuid = asm.Message.UUID
		exists = uuid.IsValid()
	}
	return
}
