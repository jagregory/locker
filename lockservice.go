package locker

import (
	"github.com/coreos/go-etcd/etcd"
	"time"
)

// Lock service client
type Client struct {
	Store Store
}

// Creates a new client using Etcd as a store
func EtcdNew(etcdclient *etcd.Client) Client {
	return Client{Store: EtcdStore{etcdclient}}
}

// Get the value of a lock, returns LockNotFound if the lock doesn't exist
func (c Client) Get(service string) (string, error) {
	return c.Store.Get(service)
}

// Lock a key and set its value
func (c Client) Lock(service, host string, stateChanges chan<- StateChange, quit <-chan bool) {
	lastState := StateChange{}
	tick := time.Tick(time.Second * 3)

	for {
		select {
		case <-quit:
			return
		case <-tick:
			state := c.updateNode(service, host)

			if stateChanges != nil && lastState.State != state.State {
				stateChanges <- state
			}

			lastState = state
		}
	}
}

// Update the lock node in the cluster.
func (c Client) updateNode(service, value string) StateChange {
	if err := c.Store.AcquireOrFreshenLock(service, value); err != nil {
		if _, ok := err.(LockDenied); ok {
			return StateChange{Lost, time.Now(), nil}
		}

		// no idea what just happened, just return the error
		return StateChange{Error, time.Now(), err}
	}

	return StateChange{Won, time.Now(), nil}
}

// A backing store for the lock service. Needs to be able to support
// querying and an atomic compare-and-swap.
type Store interface {
	// Get the value of a lock, returns LockNotFound if the lock doesn't exist
	Get(service string) (string, error)

	// Creates a lock entry if it doesn't exist, or refreshes it if it does
	AcquireOrFreshenLock(service, value string) error
}
