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

func TestConnectivity(t *testing.T) {
	client := Client{Store: EtcdStore{Etcd: etcdclient}}

	stateChanges1st := make(chan StateChange, 100)
	quit1st := make(chan bool)
	go client.Lock("srv", "1st-value", stateChanges1st, quit1st)

	time.Sleep(7 * time.Second)

	stateChanges2nd := make(chan StateChange, 100)
	quit2nd := make(chan bool)
	go client.Lock("srv", "2nd-value", stateChanges2nd, quit2nd)

	time.Sleep(7 * time.Second)

	// first client won
	changes := readStateChanges(stateChanges1st)
	if len(changes) != 1 {
		t.Errorf("Expected 1 change, got %d", len(changes))
	}

	if changes[0].State != Won {
		t.Errorf("Expected first state change to be Won, got %d", changes[0].State)
	}

	// second client lost
	changes = readStateChanges(stateChanges2nd)
	if len(changes) != 1 {
		t.Errorf("Expected 1 change, got %d", len(changes))
	}

	if changes[0].State != Lost {
		t.Errorf("Expected first state change to be Lost, got %d", changes[0].State)
	}

	// abandon the first client, so 2nd client should grab its lock
	quit1st <- true
	time.Sleep(7 * time.Second)

	changes = readStateChanges(stateChanges2nd)
	if len(changes) != 1 {
		t.Errorf("Expected 1 change, got %d", len(changes))
	}

	if changes[0].State != Won {
		t.Errorf("Expected first state change to be Won, got %d", changes[0].State)
	}

	quit2nd <- true
}

func readStateChanges(changes <-chan StateChange) []StateChange {
	s := []StateChange{}

	select {
	case c := <-changes:
		s = append(s, c)
	default:
	}

	return s
}
