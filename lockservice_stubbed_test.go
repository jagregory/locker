package locker

import "testing"

const service = "myservice"

func TestGetOnValidLock(t *testing.T) {
	expectedValue := "http://hostname"
	store := &memoryStore{}
	store.set(service, expectedValue)

	val, err := Client{store}.Get(service)
	if err != nil {
		t.Fatal(err)
	}

	if val != expectedValue {
		t.Errorf("Expected value to be '%s' got '%s'", expectedValue, val)
	}
}

func TestGetOnMissingLockReturnsLockNotFound(t *testing.T) {
	store := &memoryStore{}
	val, err := Client{store}.Get(service)
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

	stateChanges := make(chan StateChange)
	quit := make(chan bool)

	go client.Lock(service, "host", stateChanges, quit)

	select {
	case change := <-stateChanges:
		if change.State != Won {
			t.Errorf("Expected initial Won state, got %s", change.State)
		}
	}

	v, _ := store.Get(service)
	if v != "host" {
		t.Error("Expected key to be set")
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

func (c *memoryStore) Get(service string) (string, error) {
	c.ensureCache()

	if v, ok := c.cache[service]; ok {
		return v, nil
	}

	return "", LockNotFound{service}
}

func (c *memoryStore) AcquireOrFreshenLock(service, value string) error {
	c.ensureCache()

	if v, ok := c.cache[service]; ok {
		if v != value {
			return LockDenied{service}
		}
	}

	c.cache[service] = value
	return nil
}
