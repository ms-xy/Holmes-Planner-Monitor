package client

import (
	"sync"
)

type MonitorState int

func (ms MonitorState) String() string {
	switch ms {
	case StateDisconnected:
		return "disconnected"
	case StateConnecting:
		return "connecting"
	case StateConnected:
		return "connected"
	case StateDisconnecting:
		return "disconnecting"
	default:
		return "UNKNOWN-MONITORSTATE"
	}
}

const (
	// Connection state
	StateDisconnected MonitorState = iota
	StateConnecting
	StateConnected
	StateDisconnecting
)

var (
	// Mutex for monitor state management and the state itself
	monitorlock  *sync.Mutex  = &sync.Mutex{}
	monitorstate MonitorState = StateDisconnected
)

func stateTransition(oldstate, newstate MonitorState) (bool, MonitorState) {
	monitorlock.Lock()
	defer monitorlock.Unlock()

	if monitorstate != oldstate {
		return false, monitorstate
	}
	monitorstate = newstate
	return true, monitorstate
}
