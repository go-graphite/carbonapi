package main

import (
	"testing"
	"time"
)

func TestCacheExpire(t *testing.T) {

	c := &expireCache{cache: make(map[string]cacheElement)}

	sleep := make(chan bool)
	cleanerSleep = func(_ time.Duration) { <-sleep }
	done := make(chan bool)
	cleanerDone = func() { <-done }

	defer func() {
		cleanerSleep = time.Sleep
		cleanerDone = func() {}
		timeNow = time.Now
	}()

	go c.cleaner()
	t0 := time.Now()

	timeNow = func() time.Time { return t0 }

	c.set("foo", []byte("bar"), 30)
	c.set("baz", []byte("qux"), 60)
	c.set("zot", []byte("bork"), 120)

	type expireTest struct {
		key string
		ok  bool
	}

	// test expiration logic in get()

	present := []expireTest{
		{"foo", true},
		{"baz", true},
		{"zot", true},
	}

	// unexpired
	for _, p := range present {

		b, ok := c.get(p.key)

		if ok != p.ok || (ok != (b != nil)) {
			t.Errorf("expireCache: bad unexpired cache.get(%v)=(%v,%v), want %v", p.key, string(b), ok, p.ok)
		}
	}

	if len(c.keys) != 3 {
		t.Errorf("unexpired keys array length mismatch: got %d, want %d", len(c.keys), 3)
	}

	if c.totalSize != 3+3+4 {
		t.Errorf("unexpired cache size mismatch: got %d, want %d", c.totalSize, 3+3+4)
	}

	c.set("baz", []byte("snork"), 60)

	if len(c.keys) != 3 {
		t.Errorf("unexpired extra keys array length mismatch: got %d, want %d", len(c.keys), 3)
	}

	if c.totalSize != 3+5+4 {
		t.Errorf("unexpired extra cache size mismatch: got %d, want %d", c.totalSize, 3+3+4)
	}

	// expire key `foo`
	timeNow = func() time.Time { return t0.Add(45 * time.Second) }

	present = []expireTest{
		{"foo", false},
		{"baz", true},
		{"zot", true},
	}

	for _, p := range present {
		b, ok := c.get(p.key)
		if ok != p.ok || (ok != (b != nil)) {
			t.Errorf("expireCache: bad partial expire cache.get(%v)=(%v,%v), want %v", p.key, string(b), ok, p.ok)
		}
	}

	// let the cleaner run
	timeNow = func() time.Time { return t0.Add(75 * time.Second) }
	sleep <- true
	done <- true

	present = []expireTest{
		{"foo", false},
		{"baz", false},
		{"zot", true},
	}

	for _, p := range present {
		b, ok := c.get(p.key)
		if ok != p.ok || (ok != (b != nil)) {
			t.Errorf("expireCache: bad partial expire cache.get(%v)=(%v,%v), want %v", p.key, string(b), ok, p.ok)
		}
	}

	if len(c.keys) != 1 {
		t.Errorf("unexpired keys array length mismatch: got %d, want %d", len(c.keys), 3)
	}

	if c.totalSize != 4 {
		t.Errorf("unexpired cache size mismatch: got %d, want %d", c.totalSize, 3+3+4)
	}
}
