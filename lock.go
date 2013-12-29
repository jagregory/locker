package locker

import "time"

type lockState int

const (
	unknown  lockState = iota
	acquired lockState = iota
	released lockState = iota
)

// Lock a key and set its value. If the owned channel is provided, a bool will be
// pushed whenever our ownership of the lock changes. Pushing true into the quit
// channel will stop the locker and let the lock expire if we own it.
func (c Client) Lock(name, value string, owned chan<- bool, quit <-chan bool) error {
	lastState := unknown
	tick := time.Tick(time.Second * 3)

	for {
		select {
		case <-quit:
			return nil
		case <-tick:
			state, err := c.updateNode(name, value)
			if err != nil {
				return err
			}

			if owned != nil && lastState != state {
				owned <- state == acquired
			}

			lastState = state
		}
	}

	panic("unreachable")
}

// Update the lock node in the cluster.
func (c Client) updateNode(name, value string) (lockState, error) {
	if err := c.Store.AcquireOrFreshenLock(name, value); err != nil {
		if _, ok := err.(LockDenied); ok {
			return released, nil
		}

		// no idea what just happened, just return the error
		return unknown, err
	}

	return acquired, nil
}
