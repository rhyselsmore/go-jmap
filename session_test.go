package jmap

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestNewDefaultSessionCache(t *testing.T) {
	cache := NewDefaultSessionCache(5 * time.Minute)
	if cache == nil {
		t.Fatal("returned nil")
	}
}

func TestDefaultSessionCacheGet(t *testing.T) {
	t.Run("fetches on first call", func(t *testing.T) {
		cache := NewDefaultSessionCache(time.Minute)
		calls := 0
		getter := func(_ context.Context) (Session, error) {
			calls++
			return Session{isSet: true, Username: "user@example.com"}, nil
		}

		sess, err := cache.Get(context.Background(), getter)
		if err != nil {
			t.Fatalf("Get() error: %v", err)
		}
		if sess.Username != "user@example.com" {
			t.Errorf("Username = %q, want %q", sess.Username, "user@example.com")
		}
		if calls != 1 {
			t.Errorf("getter called %d times, want 1", calls)
		}
	})

	t.Run("returns cached on second call", func(t *testing.T) {
		cache := NewDefaultSessionCache(time.Minute)
		calls := 0
		getter := func(_ context.Context) (Session, error) {
			calls++
			return Session{isSet: true, Username: "user@example.com"}, nil
		}

		cache.Get(context.Background(), getter)
		cache.Get(context.Background(), getter)

		if calls != 1 {
			t.Errorf("getter called %d times, want 1", calls)
		}
	})

	t.Run("getter error propagates", func(t *testing.T) {
		cache := NewDefaultSessionCache(time.Minute)
		getter := func(_ context.Context) (Session, error) {
			return Session{}, errors.New("network error")
		}

		_, err := cache.Get(context.Background(), getter)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestDefaultSessionCacheInvalidate(t *testing.T) {
	cache := NewDefaultSessionCache(time.Minute)
	calls := 0
	getter := func(_ context.Context) (Session, error) {
		calls++
		return Session{isSet: true}, nil
	}

	cache.Get(context.Background(), getter)
	cache.Invalidate(context.Background())
	cache.Get(context.Background(), getter)

	if calls != 2 {
		t.Errorf("getter called %d times after invalidate, want 2", calls)
	}
}

func TestDefaultSessionCacheTTLExpiry(t *testing.T) {
	cache := NewDefaultSessionCache(10 * time.Millisecond)
	calls := 0
	getter := func(_ context.Context) (Session, error) {
		calls++
		return Session{isSet: true}, nil
	}

	cache.Get(context.Background(), getter)
	time.Sleep(20 * time.Millisecond)
	cache.Get(context.Background(), getter)

	if calls != 2 {
		t.Errorf("getter called %d times, want 2 (TTL expired)", calls)
	}
}

func TestDefaultSessionCacheConcurrent(t *testing.T) {
	cache := NewDefaultSessionCache(time.Minute)
	var mu sync.Mutex
	calls := 0
	getter := func(_ context.Context) (Session, error) {
		mu.Lock()
		calls++
		mu.Unlock()
		return Session{isSet: true}, nil
	}

	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := cache.Get(context.Background(), getter)
			if err != nil {
				t.Errorf("Get() error: %v", err)
			}
		}()
	}
	wg.Wait()

	if calls > 5 {
		t.Errorf("getter called %d times, expected much fewer with caching", calls)
	}
}
