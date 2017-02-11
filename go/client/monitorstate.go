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

type MonitorStateManager struct {
	// Mutex for monitor state management and the state itself
	monitorlock  *sync.Mutex
	monitorstate MonitorState
}

func (this *MonitorStateManager) stateTransition(oldstate, newstate MonitorState) (bool, MonitorState) {
	this.monitorlock.Lock()
	defer this.monitorlock.Unlock()

	if this.monitorstate != oldstate {
		return false, this.monitorstate
	}
	this.monitorstate = newstate
	return true, this.monitorstate
}
