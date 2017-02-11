package server

import (
	"crypto/sha256"
	"github.com/ms-xy/Holmes-Planner-Monitor/go/msgtypes"
	"net"
	"sync"
	"time"
)

//
// A session object represents a single planner connected to the status API
//
type Session struct {
	sync.Mutex

	// uuid 				= only available unique identifier of a planner
	// machine_uuid = unique identifier of the machine that a planner belongs to
	// pid 					= per-machine unique identifier of a planner
	uuid         *msgtypes.UUID
	machine_uuid *msgtypes.UUID
	pid          uint64

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
// Initialize fields within the session object.
//
func (c *Session) init() {
	c.kvStore = make(map[[32]byte]interface{})
}

//
// Get the session's UUID.
// Unique.
// This is afaik the planner's UUID.
//
func (c *Session) GetUuid() *msgtypes.UUID {
	return c.uuid
}

//
// Get the associated machines UUID.
// Unique per machine.
//
func (c *Session) GetMachineUuid() *msgtypes.UUID {
	return c.machine_uuid
}

//
// Get the session's PID.
// Unique per machine.
//
// func (c *Session) GetPID() uint64 {
// 	return c.pid
// }

//
// Mark a session as closed. This results in session removal by the dispatcher.
// Sessions are only removed after all their messages have been processed.
// TODO: review
//
func (c *Session) Close() {
	c._close = true
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

//
// Get session data
//
func (c *Session) GetData(key string) (interface{}, bool) {
	c.Lock()
	defer c.Unlock()
	_key := sha256.Sum256([]byte(key))
	data, exists := c.kvStore[_key]
	return data, exists
}
