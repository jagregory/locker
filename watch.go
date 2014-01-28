package locker

import "time"

// Watch will create a watch on a lock and push value changes into the
// valueChanges channel. An empty string indicates the lack of a lock.
// Pushing true into quit will end the watch.
//
//     valueChanges := make(chan string)
//
//     go client.Watch("my-service", valueChanges, nil)
//
//     for {
//       select {
//       case v := <-valueChanges:
//         fmt.Printf("Lock value changed: %s\n", v)
//       }
//     }
//
// Watch is a blocking call, so it's recommended to run it in a goroutine.
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
