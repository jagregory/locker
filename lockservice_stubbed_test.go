package locker

import "testing"

const name = "myservice"

func TestGetOnValidLock(t *testing.T) {
	expectedValue := "http://hostname"
	store := &memoryStore{}
	store.set(name, expectedValue)

	val, err := Client{store}.Get(name)
	if err != nil {
		t.Fatal(err)
	}

	if val != expectedValue {
		t.Errorf("Expected value to be '%s' got '%s'", expectedValue, val)
	}
}

func TestGetOnMissingLockReturnsLockNotFound(t *testing.T) {
	store := &memoryStore{}
	val, err := Client{store}.Get(name)
	if err == nil {
		t.Error("Expected an error, didn't get one")
	}

	if _, ok := err.(LockNotFound); !ok {
		t.Errorf("Expected lock not found, got different error %s", err)
	}

	if val != "" {
		t.Errorf("Expected no value, got '%s'", val)
	}
}

func TestLockSetsKeyValue(t *testing.T) {
	store := &memoryStore{}
	client := Client{store}

	owned := make(chan bool)
	quit := make(chan bool)

	go client.Lock(name, "host", owned, quit)

	select {
	case change := <-owned:
		if change != true {
			t.Errorf("Expected initial Won state, got %s", change)
		}
	}

	v, _ := store.Get(name)
	if v != "host" {
		t.Error("Expected key to be set")
	}

	quit <- true
}

func TestWatchReturnsInitialValue(t *testing.T) {
	store := &memoryStore{}
	client := Client{store}

	valueChanges := make(chan string)
	quit := make(chan bool)

	store.set(name, "value")

	go client.Watch(name, valueChanges, quit)

	select {
	case <-timeout():
		t.Fatal("Timeout")
	case change := <-valueChanges:
		if change != "value" {
			t.Error("Expected value to be 'value'")
		}
	}

	quit <- true
}

func TestWatch(t *testing.T) {
	store := &memoryStore{}
	client := Client{store}

	valueChanges := make(chan string)
	quit := make(chan bool)

	go client.Watch(name, valueChanges, quit)

	select {
	case <-timeout():
		t.Fatal("Timeout")
	case change := <-valueChanges:
		if change != "" {
			t.Error("Expected initial blank value inidicating missing lock")
		}
	}

	store.set(name, "value")

	select {
	case <-timeout():
		t.Fatal("Timeout")
	case change := <-valueChanges:
		if change != "value" {
			t.Error("Expected value to change to 'value'")
		}
	}

	quit <- true
}

// For testing purposes
type memoryStore struct {
	cache map[string]string
}

func (c *memoryStore) ensureCache() {
	if c.cache == nil {
		c.cache = make(map[string]string)
	}
}

func (c *memoryStore) set(key, value string) {
	c.ensureCache()
	c.cache[key] = value
}

func (c *memoryStore) Get(name string) (string, error) {
	c.ensureCache()

	if v, ok := c.cache[name]; ok {
		return v, nil
	}

	return "", LockNotFound{name}
}

func (c *memoryStore) AcquireOrFreshenLock(name, value string) error {
	c.ensureCache()

	if v, ok := c.cache[name]; ok {
		if v != value {
			return LockDenied{name}
		}
	}

	c.cache[name] = value
	return nil
}
