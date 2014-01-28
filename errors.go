package locker

import "fmt"

// LockNotFound is returned when a Get is made against a key which isn't
// locked.
type LockNotFound struct {
	service string
}

func (e LockNotFound) Error() string {
	return fmt.Sprintf("Lock not found: %s", e.service)
}

// LockDenied is returned when a lock is attempted on a key which has been
// acquired by another client.
type LockDenied struct {
	key string
}

func (e LockDenied) Error() string {
	return fmt.Sprintf("Lock attempt was denied: %s", e.key)
}
