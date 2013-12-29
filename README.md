# locker

A distributed lock service client for [etcd](https://github.com/coreos/etcd).

## Usage

### Creating a lock

```go
import "github.com/jagregory/locker"

// Create the locker client.
// Etcd is needed as a backend, create your etcd client elsewhere and pass it in.
client := locker.New(etcdclient)

go client.Lock("key", "value", nil, nil)

client.Get("key") // "value"
```

### Listening for state changes

The third argument to `Lock` is a `owned` channel. When a transition of ownership occurs a bool will be dumped into this channel indicating whether you own the lock.

```go
owned := make(chan bool)
go client.Lock("key", "value", owned, nil)

select {
case change := <-owned:
	fmt.Printf("Do we own the lock? %b", change)
}
```

Pass `nil` for this channel if you don't intend to read from it. Not doing so will block the lock from refreshing.

### Releasing the lock

The fourth argument to `Lock` is a `quit` channel. Push anything into this channel to kill the locking, which will let the lock expire.

```go
quit := make(chan bool)
go client.Lock("key", "value", nil, quit)

quit <- true
```

### Watching a lock

An interesting aspect of lock services is the ability to watch a lock that isn't owned by you. A service can alter its behaviour depending on the value of a lock. You can use the `Watch` function to watch the value of a lock.

```go
valueChanges := make(chan string)
go client.Watch("key", valueChanges, nil)

select {
case change := <-valueChanges
	fmt.Printf("Lock value changed: %s", change)
}
```

Quitting works the same way as `Lock`.

## Gotchas

The `Lock` call is blocking, use a goroutine.

The `Lock` method takes a `owned` channel. Pass `nil` for this if you don't intend to read from it. Passing a channel and not reading from it will block the refreshing of the lock and you'll lose it.