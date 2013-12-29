package locker

import "github.com/coreos/go-etcd/etcd"

// Lock service client
type Client struct {
	Store Store
}

// Creates a new client using Etcd as a store
func New(etcdclient *etcd.Client) Client {
	return Client{
		Store: EtcdStore{
			Etcd: etcdclient,
		},
	}
}

// Get the value of a lock, returns LockNotFound if the lock doesn't exist
func (c Client) Get(name string) (string, error) {
	return c.Store.Get(name)
}

// A backing store for the lock service. Needs to be able to support
// querying and an atomic compare-and-swap.
type Store interface {
	// Get the value of a lock, returns LockNotFound if the lock doesn't exist
	Get(name string) (string, error)

	// Creates a lock entry if it doesn't exist, or refreshes it if it does
	AcquireOrFreshenLock(name, value string) error
}
