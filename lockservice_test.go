package locker

import (
	_ "fmt"
	"github.com/coreos/go-etcd/etcd"
	"testing"
	"time"
)

var etcdclient *etcd.Client

func init() {
	etcdclient = etcd.NewClient(nil)
	_, err := etcdclient.Get("test", false, false)
	if err != nil && err.Error() == "Cannot reach servers after 3 time" {
		panic("Can't run tests without an etcd instance running")
	}
}

func TestTwoCompetingClients(t *testing.T) {
	client := Client{Store: EtcdStore{Etcd: etcdclient}}

	// create a lock using one client
	owned1st := make(chan bool)
	quit1st := make(chan bool)
	go client.Lock("srv", "1st-value", owned1st, quit1st)

	// verify we acquired the lock
	select {
	case <-timeout():
		t.Fatal("Timeout")
	case v := <-owned1st:
		if v == false {
			t.Fatal("Expected first client to acquire the lock")
		}
	}

	// try to acquire the lock with a second client
	owned2nd := make(chan bool)
	quit2nd := make(chan bool)
	go client.Lock("srv", "2nd-value", owned2nd, quit2nd)

	// verify we failed to do that
	select {
	case <-timeout():
		t.Fatal("Timeout")
	case v := <-owned2nd:
		if v == true {
			t.Fatal("Expected second client to not acquire the lock")
		}
	}

	// abandon the first client
	quit1st <- true

	// verify 2nd client now acquired the lock
	select {
	case <-timeout():
		t.Fatal("Timeout")
	case v := <-owned2nd:
		if v == false {
			t.Fatal("Expected second client to acquire the lock after the first releases it")
		}
	}

	quit2nd <- true
}

func timeout() <-chan time.Time {
	return time.After(10 * time.Second)
}
