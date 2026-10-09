package communication

import (
	"fmt"
	"sync"

	"datalogger/internal/logger"
)

// StateChangeCallback is invoked whenever a device connection state changes
type StateChangeCallback func(deviceID uint, from, to ConnectionState, reason string)

// ConnectionStateMachine manages deterministic lifecycle state transitions for edge devices
type ConnectionStateMachine struct {
	mu        sync.RWMutex
	states    map[uint]ConnectionState
	callbacks []StateChangeCallback
}

// NewConnectionStateMachine constructs a new thread-safe state machine
func NewConnectionStateMachine() *ConnectionStateMachine {
	return &ConnectionStateMachine{
		states: make(map[uint]ConnectionState),
	}
}

// OnStateChange registers an observer callback for state transitions
func (sm *ConnectionStateMachine) OnStateChange(cb StateChangeCallback) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks = append(sm.callbacks, cb)
}

// RegisterCallback is an alias for OnStateChange
func (sm *ConnectionStateMachine) RegisterCallback(cb StateChangeCallback) {
	sm.OnStateChange(cb)
}

// GetState returns the current connection state of a device (defaults to DISCONNECTED)
func (sm *ConnectionStateMachine) GetState(deviceID uint) ConnectionState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if s, exists := sm.states[deviceID]; exists {
		return s
	}
	return StateDisconnected
}

// SetInitialState sets the state without transition validation (used during initial discovery)
func (sm *ConnectionStateMachine) SetInitialState(deviceID uint, s ConnectionState) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.states[deviceID] = s
}

// CanTransition validates whether moving from 'from' to 'to' is allowed
func CanTransition(from, to ConnectionState) bool {
	if from == to {
		return true
	}

	switch from {
	case StateDisconnected:
		// From DISCONNECTED: can start CONNECTING or be marked DISABLED
		return to == StateConnecting || to == StateDisabled

	case StateConnecting:
		// From CONNECTING: can succeed (CONNECTED), fail (ERROR), disconnect (DISCONNECTED), or disable
		return to == StateConnected || to == StateError || to == StateDisconnected || to == StateDisabled

	case StateConnected:
		// From CONNECTED: can degrade on threshold failures (DEGRADED), error, disconnect, or disable
		return to == StateDegraded || to == StateDisconnected || to == StateError || to == StateDisabled

	case StateDegraded:
		// From DEGRADED: can attempt reconnect (RECONNECTING), recover (CONNECTED), error out (ERROR), disconnect, or disable
		return to == StateReconnecting || to == StateConnected || to == StateError || to == StateDisconnected || to == StateDisabled

	case StateReconnecting:
		// From RECONNECTING: can succeed (CONNECTED), retry in backoff (DEGRADED), exceed max (ERROR), disconnect, or disable
		return to == StateConnected || to == StateDegraded || to == StateError || to == StateDisconnected || to == StateDisabled

	case StateError:
		// From ERROR: can retry (RECONNECTING / CONNECTING), disconnect, or disable
		return to == StateReconnecting || to == StateConnecting || to == StateDisconnected || to == StateDisabled

	case StateDisabled:
		// From DISABLED: can only transition to DISCONNECTED when re-enabled
		return to == StateDisconnected

	default:
		return false
	}
}

// Transition attempts to transition the device to a new state with validation and notification
func (sm *ConnectionStateMachine) Transition(deviceID uint, to ConnectionState, reason string) (ConnectionState, error) {
	sm.mu.Lock()
	currentState, exists := sm.states[deviceID]
	if !exists {
		currentState = StateDisconnected
	}

	if currentState == to {
		sm.mu.Unlock()
		return currentState, nil
	}

	if !CanTransition(currentState, to) {
		sm.mu.Unlock()
		return currentState, fmt.Errorf("invalid connection state transition from %s to %s for device %d (reason: %s)",
			currentState, to, deviceID, reason)
	}

	sm.states[deviceID] = to
	callbacks := make([]StateChangeCallback, len(sm.callbacks))
	copy(callbacks, sm.callbacks)
	sm.mu.Unlock()

	logger.Debug("Device %d connection state changed: %s -> %s (reason: %s)", deviceID, currentState, to, reason)

	for _, cb := range callbacks {
		if cb != nil {
			cb(deviceID, currentState, to, reason)
		}
	}

	return to, nil
}

// RemoveDevice cleans up state memory when a device is deleted
func (sm *ConnectionStateMachine) RemoveDevice(deviceID uint) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.states, deviceID)
}
