# locker

A distributed lock service client for [etcd](https://github.com/coreos/etcd).

## Usage

```go
import "github.com/jagregory/locker"

client := locker.Client{
	Store: locker.EtcdStore{etcdclient},
}

go client.Lock("my-service", "my-hostname", nil, nil)

client.Get("my-service") // "my-hostname"
```

## Listening for state changes

The third argument to `Lock` is a `stateChanges` channel. When a state transition occurs a message will be dumped into this channel.

```go
stateChanges := make(chan locker.StateChange)
go client.Lock("my-service", "my-hostname", stateChanges, nil)

select {
case change := <- stateChanges:
	fmt.Print("State changed")
}
```

Each `StateChange` has a `State` value and an `Error` if one occurred.

Pass `nil` for this channel if you don't intend to read from it. Not doing so will block the lock from refreshing.

## Releasing the lock

The fourth argument to `Lock` is a `quit` channel. Push anything into this channel to kill the locking, which will let the lock expire.

```go
quit := make(chan bool)
go client.Lock("my-service", "my-hostname", nil, quit)

quit <- true
```

## Gotchas

The `Lock` call is blocking, use a goroutine.

The `Lock` method takes a `stateChanges` channel. Pass `nil` for this if you don't intend to read from it. Passing a channel and not reading from it will block the refreshing of the lock and you'll lose it.