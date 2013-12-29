package locker

import "time"

// Watch a lock and be notified when its value changes. An empty string indicates the
// lack of a lock. Pushing true into quit will end the watch.
func (c Client) Watch(name string, valueChanges chan<- string, quit <-chan bool) error {
	var lastValue string
	first := true

	for {
		select {
		case <-quit:
			return nil
		default:
			v, err := c.Store.Get(name)

			if err != nil {
				if _, ok := err.(LockNotFound); ok {
					v = ""
				} else {
					return err
				}
			}

			if v != lastValue || first {
				valueChanges <- v
				lastValue = v
			}

			first = false
			time.Sleep(3 * time.Second)
		}
	}
}
