package locker

import "github.com/coreos/go-etcd/etcd"

const ttl uint64 = 5

type EtcdStore struct {
	Etcd *etcd.Client
}

func (s EtcdStore) Get(service string) (string, error) {
	res, err := s.Etcd.Get("ns/"+service, false, false)
	if err == nil {
		return res.Node.Value, nil
	}

	if etcderr, ok := err.(etcd.EtcdError); ok && etcderr.ErrorCode == 100 {
		return "", LockNotFound{service}
	}

	return "", err
}

func (s EtcdStore) AcquireOrFreshenLock(service, value string) error {
	key := "ns/" + service
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
			return s.AcquireOrFreshenLock(service, value)
		case 101:
			// couldn't set the key, the prevValue we gave it differs from the
			// one in the server. Someone else has this key.
			return LockDenied{service}
		}
	}

	return err
}
