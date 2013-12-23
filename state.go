package locker

import "time"

type State int

const (
	_ = iota

	// An error occurred whilst trying to establish or maintain a lock
	Error State = iota

	// A lock attempt was lost, someone else got to it before us
	Lost State = iota

	// A lock attempt was won, we acquired the lock
	Won State = iota
)

// A change of state of the lock, Won->Lost Lost->Won, and when it occurred.
type StateChange struct {
	// New state of the lock
	State State

	// Time the state change occurred
	Time time.Time

	// The error, if one occurred as part of changing state
	Error error
}
