package locker

import "fmt"

type LockNotFound struct {
	service string
}

func (e LockNotFound) Error() string {
	return fmt.Sprintf("Lock not found: %s", e.service)
}

type LockDenied struct {
	key string
}

func (e LockDenied) Error() string {
	return fmt.Sprintf("Lock attempt was denied: %s", e.key)
}
