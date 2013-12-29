package locker

import "github.com/coreos/go-etcd/etcd"

const ttl uint64 = 5

type EtcdStore struct {
	// Etcd client used for storing locks
	Etcd *etcd.Client

	// Directory in Etcd to store locks. Default: locker
	Directory string
}

// Gets the value of a lock. Returns LockNotFound if a lock with the name isn't held
func (s EtcdStore) Get(name string) (string, error) {
	res, err := s.Etcd.Get(s.lockPath(name), false, false)
	if err == nil {
		return res.Node.Value, nil
	}

	if etcderr, ok := err.(etcd.EtcdError); ok && etcderr.ErrorCode == 100 {
		return "", LockNotFound{name}
	}

	return "", err
}

// Aquires a named lock, or updates its TTL if it is already held
func (s EtcdStore) AcquireOrFreshenLock(name, value string) error {
	key := s.lockPath(name)
	_, err := s.Etcd.CompareAndSwap(key, value, ttl, value, 0)
	if err == nil {
		// success!
		return nil
	}

	if etcderr, ok := err.(etcd.EtcdError); ok {
		switch etcderr.ErrorCode {
		case 100:
			// key doesn't exist, set it. This seems to be odd behaviour for
			// CompareAndSwap. Surely, if it doesn't exist we should just set
			// it as part of CompareAndSwap. Potential for a race condition here,
			// where another client is able to do a CompareAndSwap and then we
			// stomp on it with our dumb Set.
			if _, err := s.Etcd.Set(key, value, 1); err != nil {
				// wasn't able to force-set the key, no idea what happened
				return err
			}

			// Retry after stomping
			return s.AcquireOrFreshenLock(name, value)
		case 101:
			// couldn't set the key, the prevValue we gave it differs from the
			// one in the server. Someone else has this key.
			return LockDenied{name}
		}
	}

	return err
}

// Gets the path to a lock in Etcd
func (s EtcdStore) lockPath(name string) string {
	d := s.Directory
	if d == "" {
		d = "locker"
	}

	return d + "/" + name
}
