package server

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	// "github.com/ms-xy/Holmes-Planner-Monitor/go/msgtypes"
	"net"
	"sync"
	"time"
)

//
// A session object represents a single client connected to the status API
//
type Session struct {
	sync.Mutex

	id            uint64
	addrKey       [18]byte
	listenAddrKey [18]byte

	address       *net.UDPAddr
	listenAddress *net.TCPAddr // TODO: implement in all the functions (especially parsers)
	_close        bool

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
func (c *Session) GetID() uint64 {
	return c.id
}

func (c *Session) GetAddress() *net.UDPAddr {
	return c.address
}

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
		id_next:           1,
		map_addr2id:       make(map[[18]byte]uint64),
		map_listenAddr2id: make(map[[18]byte]uint64),
		map_id2session:    make(map[uint64]*Session),
	}
}

type SessionMap struct {
	sync.Mutex
	id_next           uint64
	map_addr2id       map[[18]byte]uint64
	map_listenAddr2id map[[18]byte]uint64
	map_id2session    map[uint64]*Session
}

// Check whether a session exists for the client. If not create a new one.
// Returns the session object as well as a boolean indicating if the client is
// new.
func (this *SessionMap) StartSession(asm AddressedStatusMessage) (*Session, bool) {
	this.Lock()
	defer this.Unlock()
	var (
		session *Session
		id      uint64
		exists  bool
	)
	// calculate the master key first
	addrKey := addr2bytemap18key(asm.Address.IP, asm.Address.Port)
	listenAddrKey, lak_provided := extractListenAddrKey(asm)
	if lak_provided && !bytemap18key_IP_equal(addrKey, listenAddrKey) {
		// TODO: potential fatal error, node may be pretending to be someone else? (misconfiguration? etc?)
	}

	if id, exists = this.map_addr2id[addrKey]; exists {
		session = this.map_id2session[id]
		if lak_provided && !bytemap18key_equal(session.listenAddrKey, listenAddrKey) {
			// update potential old entry in listenAddrKey map
			delete(this.map_listenAddr2id, session.listenAddrKey)
			this.map_listenAddr2id[listenAddrKey] = session.id
		}

	} else if lak_provided {
		if id, exists = this.map_listenAddr2id[listenAddrKey]; exists {
			session = this.map_id2session[id]
			// update potential old entry in addrKey map
			if !bytemap18key_equal(session.addrKey, addrKey) {
				delete(this.map_addr2id, addrKey)
				this.map_addr2id[addrKey] = session.id
			}
		}
	}

	if session == nil {
		fmt.Println("-- new addrKey and listenAddrKey")
		fmt.Println("\t", addrKey)
		fmt.Println("\t", listenAddrKey)
		id = this.id_next
		this.id_next++
		// create a new session object as the client seems to be unknown
		session = &Session{}
		this.map_addr2id[addrKey] = id
		this.map_id2session[id] = session
		// update lak map
		if lak_provided {
			this.map_listenAddr2id[listenAddrKey] = id
		}
	}

	session.id = id
	session.addrKey = addrKey
	session.listenAddrKey = listenAddrKey

	return session
}

// Remove a session from the session storage
func (this *SessionMap) Remove(session *Session) {
	this.Lock()
	this.Unlock()
	delete(this.map_id2session, session.id)
	delete(this.map_addr2id, session.addrKey)
	delete(this.map_listenAddr2id, session.listenAddrKey)
}

// Get a session by its ID
func (this *SessionMap) Get(id uint64) (*Session, bool) {
	session, exists := this.map_id2session[id]
	return session, exists
}

// Loop over all sessions in the session storage and execute the function fn for
// every single one
func (this *SessionMap) ForEach(fn func(*Session)) {
	this.Lock()
	defer this.Unlock()
	for _, s := range this.map_id2session {
		fn(s)
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

func extractListenAddrKey(asm AddressedStatusMessage) ([18]byte, bool) {
	var listenAddrKey [18]byte
	if asm.Message.PlannerInfo != nil && asm.Message.PlannerInfo.ListenAddress != "" {
		listenAddr, _ := net.ResolveTCPAddr("tcp", asm.Message.PlannerInfo.ListenAddress)
		listenAddrKey = addr2bytemap18key(listenAddr.IP, listenAddr.Port)
		return listenAddrKey, true
	}
	return listenAddrKey, false
}
