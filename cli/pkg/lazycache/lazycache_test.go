package lazycache_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"hmans.de/chatto/pkg/lazycache"
)

func TestGetOrCreate_Basic(t *testing.T) {
	c := lazycache.New[string]()

	val, err := c.GetOrCreate("key", func() (string, error) {
		return "hello", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "hello" {
		t.Fatalf("got %q, want %q", val, "hello")
	}

	// Second call should return cached value without calling create.
	val, err = c.GetOrCreate("key", func() (string, error) {
		t.Fatal("create should not be called for cached key")
		return "", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "hello" {
		t.Fatalf("got %q, want %q", val, "hello")
	}
}

func TestGetOrCreate_Error(t *testing.T) {
	c := lazycache.New[int]()
	wantErr := errors.New("creation failed")

	_, err := c.GetOrCreate("key", func() (int, error) {
		return 0, wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("got error %v, want %v", err, wantErr)
	}

	// After an error, the key should not be cached — a subsequent call
	// should invoke create again.
	val, err := c.GetOrCreate("key", func() (int, error) {
		return 42, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 42 {
		t.Fatalf("got %d, want 42", val)
	}
}

func TestGetOrCreate_Concurrent(t *testing.T) {
	c := lazycache.New[int]()
	var calls atomic.Int64

	var wg sync.WaitGroup
	const goroutines = 100

	for i := range goroutines {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			val, err := c.GetOrCreate("shared", func() (int, error) {
				calls.Add(1)
				return 99, nil
			})
			if err != nil {
				t.Errorf("goroutine %d: unexpected error: %v", i, err)
			}
			if val != 99 {
				t.Errorf("goroutine %d: got %d, want 99", i, val)
			}
		}(i)
	}

	wg.Wait()

	// Due to double-checked locking, create may be called more than once
	// if multiple goroutines pass the read-lock check before any acquires
	// the write lock. But it should be a small number, not 100.
	if n := calls.Load(); n > 5 {
		t.Errorf("create called %d times, expected at most a few", n)
	}
}

func TestGet(t *testing.T) {
	c := lazycache.New[string]()

	_, ok := c.Get("missing")
	if ok {
		t.Fatal("expected Get to return false for missing key")
	}

	c.Set("present", "value")
	val, ok := c.Get("present")
	if !ok {
		t.Fatal("expected Get to return true for present key")
	}
	if val != "value" {
		t.Fatalf("got %q, want %q", val, "value")
	}
}

func TestSetOverwrite(t *testing.T) {
	c := lazycache.New[int]()
	c.Set("key", 1)
	c.Set("key", 2)

	val, ok := c.Get("key")
	if !ok || val != 2 {
		t.Fatalf("got (%d, %v), want (2, true)", val, ok)
	}
}

func TestDelete(t *testing.T) {
	c := lazycache.New[string]()
	c.Set("key", "value")
	c.Delete("key")

	_, ok := c.Get("key")
	if ok {
		t.Fatal("expected key to be deleted")
	}

	// Deleting a non-existent key should not panic.
	c.Delete("nonexistent")
}

func TestMultipleKeys(t *testing.T) {
	c := lazycache.New[int]()

	for i := range 10 {
		key := string(rune('a' + i))
		val, err := c.GetOrCreate(key, func() (int, error) {
			return i, nil
		})
		if err != nil {
			t.Fatalf("key %s: unexpected error: %v", key, err)
		}
		if val != i {
			t.Fatalf("key %s: got %d, want %d", key, val, i)
		}
	}

	// Verify all keys are independently cached.
	for i := range 10 {
		key := string(rune('a' + i))
		val, ok := c.Get(key)
		if !ok {
			t.Fatalf("key %s: not found", key)
		}
		if val != i {
			t.Fatalf("key %s: got %d, want %d", key, val, i)
		}
	}
}
