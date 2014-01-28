# locker

[![GoDoc](https://godoc.org/github.com/jagregory/locker?status.png)](https://godoc.org/github.com/jagregory/locker)

A distributed lock service client for [etcd](https://github.com/coreos/etcd).

## What? Why?

A distributed lock service is somewhat self-explanatory. Locking (mutexes) as a service that's distributed across a cluster.

What a lock service gives you is a key-value store with locking semantics, and a mechanism for watching state changes. The distributed part of a distributed lock service is thanks to the etcd underpinnings, and it simply means your locks are persisted across a cluster.

A couple of examples of what you can build with one of these services:

### Name server

You could use a lock service just like a DNS, except a DNS that's populated by the hosts it fronts rather than being an independent store.

A service would lock its hostname record and set the value to its IP address:

```go
client.Lock("www.example.com", "10.1.1.1", nil, nil)
```

Then clients wanting to resolve the address could `client.Get("www.example.com")` and recieve `10.1.1.1`.

This gets a bit more interesting when clients start watching the records and are notified of changes when they occur.

```go
valueChanges := make(chan string)
go client.Watch("www.example.com", valueChanges, nil)

select {
case newIp := <-valueChanges:
	redirectRequestsToNewIp(newIp)
}
```

### Mesh/grid health monitoring

Imagine a grid or mesh of services which are all polling each-other for updates. When one service goes down, we don't want the others continuing to hammer it until it's in a healthy state again. You can do this quite easily with a lock service.

Each service locks a well-known record of its own when it's healthy.

```go
// in service-a
client.Lock("service-a", "ok", nil, quit)

// in service-b
client.Lock("service-b", "ok", nil, quit)
```

When it becomes unhealthy, it kills its own lock.

```go
// in service-b
quit <- true
```

Meanwhile, other services are watching the records of the services they depend on, and are notified when the lock dies. They can then respond to that service becoming unavailable.

```go
// in service-a
client.Watch("service-b", changes)

select {
case change := <-changes:
	if change == "" {
		handleServiceBGoingOffline()
	}
}
```

You can take this further by making service-a then kill its own lock, leading to a cascading blackout of sorts (if that's what you want). An example of this could be when someone cuts your internet pipe, all your services gracefully shut down until the pipe comes back and then they restart.

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
