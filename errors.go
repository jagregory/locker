package locker

import "fmt"

type LockNotFound struct {
	service string
}

func (e LockNotFound) Error() string {
	return fmt.Sprintf("Lock not found: %s", e.service)
}

type KeyNotFound struct {
	key string
}

func (e KeyNotFound) Error() string {
	return fmt.Sprintf("Key not found: %s", e.key)
}

type LockDenied struct {
	key string
}

func (e LockDenied) Error() string {
	return fmt.Sprintf("Lock attempt was denied: %s", e.key)
}
