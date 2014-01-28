// Package locker is a distributed lock service built on top of etcd.
// https://github.com/coreos/etcd
//
// Locker gives you a way of creating and maintaining locks across
// a network, useful for synchronising access to resources when multiple
// machines are involved.
//
// A simple example is a name server. A service would lock its hostname
// record and set the value to its IP address:
//
//     client.Lock("www.example.com", "10.1.1.1", nil, nil)
//
// Other clients wanting to resolve the address could call
//
//     client.Get("www.example.com")
//
// and would recieve the IP address: 10.1.1.1.
//
// This gets a bit more interesting when clients start watching the
// records and are notified of changes when they occur.
//
//     valueChanges := make(chan string)
//     go client.Watch("www.example.com", valueChanges, nil)
//
//     select {
//     case newIp := <-valueChanges:
//         redirectRequestsToNewIp(newIp)
//     }
package locker
