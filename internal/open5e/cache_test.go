package open5e

import (
	"testing"
	"time"
)

func TestCacheExpiryWithInjectedClock(t *testing.T) {
	currentTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	fakeClock := func() time.Time { return currentTime }

	cache := NewCacheWithClock(5*time.Minute, fakeClock)
	cache.Add("goblin", []byte(`{"name":"Goblin"}`))

	val, ok := cache.Get("goblin")
	if !ok {
		t.Fatalf("expected goblin to be in cache before TTL")
	}
	if string(val) != `{"name":"Goblin"}` {
		t.Fatalf("unexpected cached value: %s", string(val))
	}

	currentTime = currentTime.Add(6 * time.Minute)

	if _, ok := cache.Get("goblin"); ok {
		t.Fatalf("expected goblin to be expired after TTL")
	}
}
