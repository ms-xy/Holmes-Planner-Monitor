package server

import (
	"crypto/sha256"
	"encoding/binary"
	"github.com/ms-xy/Holmes-Planner-Monitor/go/msgtypes"
	"net"
	"sync"
	"time"
)

//
// A session object represents a single client connected to the status API
//
type Session struct {
	sync.Mutex

	// id      uint64
	// addrKey [18]byte
	uuid *msgtypes.UUID

	Address *net.UDPAddr
	// ListenAddress *net.TCPAddr
	_close bool

	FirstSeen time.Time
	LastSeen  time.Time
	Last      struct {
		SystemStatus  time.Time
		NetworkStatus time.Time
		PlannerStatus time.Time
		ServiceStatus time.Time
	}

	kvStore map[[32]byte]interface{}
}

//
// Initialize fields within the session object
//
func (c *Session) init() {
	c.kvStore = make(map[[32]byte]interface{})
}

//
// Getters for some of the non public fields
//
// func (c *Session) GetID() uint64 {
// 	return c.id
// }

func (c *Session) GetUuid() *msgtypes.UUID {
	return c.uuid
}

// func (c *Session) GetAddrKey() [18]byte {
// 	r := [18]byte{}
// 	copy(r[:], c.addrKey[:])
// 	return r
// }

//
// Close a session (dispatcher will enforce a removal in that case)
//
func (c *Session) Close() {
	c._close = true
	// TODO: remove from session map, do that in the dispatcher loop maybe?
}

//
// Modify session data
//
func (c *Session) SetData(key string, data interface{}) (interface{}, bool) {
	c.Lock()
	defer c.Unlock()
	_key := sha256.Sum256([]byte(key))
	olddata, exists := c.kvStore[_key]
	c.kvStore[_key] = data
	return olddata, exists
}

func (c *Session) GetData(key string) (interface{}, bool) {
	c.Lock()
	defer c.Unlock()
	_key := sha256.Sum256([]byte(key))
	data, exists := c.kvStore[_key]
	return data, exists
}

//
// Session map object
// Contains useful functions for operations on the client cache
//
func NewSessionMap() *SessionMap {
	return &SessionMap{
		// TODO: what to do if we *ever* reach maximum uint64? is that of any
		// concern even?
		// id_next: 1,
		// map_addr2id:    make(map[[18]byte]uint64),
		// map_uuid2id:    make(map[msgtypes.UUID]uint64),
		// map_id2session: make(map[uint64]*Session),
		map_uuid2session: make(map[msgtypes.UUID]*Session),
	}
}

type SessionMap struct {
	sync.Mutex
	id_next          uint64
	map_uuid2session map[msgtypes.UUID]*Session
	// map_addr2id      map[[18]byte]uint64
	// map_uuid2id      map[msgtypes.UUID]uint64
	// map_id2session   map[uint64]*Session
}

// Check whether a session exists for the client. If not create a new one.
// Returns the session object as well as a boolean indicating if the client is
// new.
func (this *SessionMap) StartSession(asm AddressedStatusMessage) (*Session, bool) {
	this.Lock()
	defer this.Unlock()

	var (
	// session *Session
	// id      uint64
	// exists  bool
	// err     error
	)

	// since we abandoned the multiple key approach, we can now simply check if
	// we already know the uuid or not, previously we had to rely on the IP as
	// well as a second identifier
	uuid := asm.Message.UUID
	// if the supplied uuid is not valid, recreate it automatically
	if !uuid.IsValid() {
		var err error
		for {
			uuid, err = msgtypes.UUID4()
			if err != nil {
				continue // TODO: better solution here? should we panic?
			}
			if _, exists := this.map_uuid2session[*uuid]; !exists {
				break
			}
		}
	}
	session, exists := this.map_uuid2session[*uuid]
	if !exists {
		session = &Session{
			uuid: uuid,
		}
		this.map_uuid2session[*uuid] = session
	}

	// calculate both keys
	// addrKey := addr2bytemap18key(asm.Address.IP, asm.Address.Port)
	// uuid, uuid_provided := extractUuid(asm)

	// check uuid first, it is the stronger identifier
	// if uuid_provided {
	// 	if id, exists = this.map_uuid2id[uuid]; exists {
	// 		session = this.map_id2session[id]
	// 		// if an old entry exists, eliminate
	// 		if !bytemap18key_equal(session.addrKey, addrKey) {
	// 			delete(this.map_addr2id, addrKey)
	// 			this.map_addr2id[addrKey] = session.id
	// 		}
	// 	}

	// } else if id, exists = this.map_addr2id[addrKey]; exists {
	// 	session = this.map_id2session[id]
	// 	uuid = session.uuid
	// }

	// if !exists {
	// 	id = this.id_next
	// 	this.id_next++
	// 	// if no uuid is given, create a new uuid
	// 	if !uuid_provided {
	// 		uuid, err = msgtypes.Uuid4()
	// 		if err != nil {
	// 			panic(err) // fatal error, cannot read from rand.Reader
	// 			// TODO: do this differently maybe?
	// 			// A panic here kills the dispatcher loop ... not very "gracefully"
	// 		}
	// 	}
	// 	// create a new session object as the client seems to be unknown
	// 	session = &Session{}
	// 	this.map_addr2id[addrKey] = id
	// 	this.map_uuid2id[uuid] = id
	// 	this.map_id2session[id] = session
	// }

	// session.id = id
	// session.uuid = uuid
	// session.addrKey = addrKey
	session.Address = asm.Address

	return session, !exists
}

// Remove a session from the session storage
func (this *SessionMap) Remove(session *Session) {
	this.Lock()
	this.Unlock()
	// delete(this.map_id2session, session.id)
	// delete(this.map_addr2id, session.addrKey)
	// delete(this.map_uuid2id, session.uuid)
	delete(this.map_uuid2session, *session.uuid)
}

// Get a session by its ID
// func (this *SessionMap) Get(id uint64) (*Session, bool) {
// 	session, exists := this.map_id2session[id]
// 	return session, exists
// }

// Get a session by its uuid
func (this *SessionMap) GetByUuid(uuid *msgtypes.UUID) (*Session, bool) {
	session, exists := this.map_uuid2session[*uuid]
	return session, exists
	// if id, exists := this.map_uuid2id[uuid]; exists {
	// 	return this.Get(id)
	// }
	// return nil, false
}

// Loop over all sessions in the session storage and execute the function fn for
// every single one
func (this *SessionMap) ForEach(fn func(*Session)) {
	this.Lock()
	defer this.Unlock()
	for _, s := range this.map_uuid2session {
		fn(s)
	}
}

// Get the size of the session object map
func (this *SessionMap) Size() int {
	return len(this.map_uuid2session)
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
