package open5e

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientGetMonsterDecodingAndCache(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if r.URL.Path != "/monsters/goblin/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"slug": "goblin",
			"name": "Goblin",
			"size": "Small",
			"type": "humanoid",
			"armor_class": 15,
			"hit_points": 7,
			"challenge_rating": "1/4",
			"cr": 0.25,
			"actions": [{"name": "Scimitar", "attack_bonus": 4, "damage_dice": "1d6+2"}]
		}`))
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, 2*time.Second, 5*time.Minute)

	m, err := client.GetMonster("goblin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name != "Goblin" || m.HitPoints != 7 || m.ArmorClass != 15 || len(m.Actions) != 1 {
		t.Fatalf("unexpected monster decoded: %+v", m)
	}

	_, err = client.GetMonster("goblin")
	if err != nil {
		t.Fatalf("unexpected error on cached call: %v", err)
	}
	if hits != 1 {
		t.Fatalf("expected 1 HTTP request thanks to cache, got %d", hits)
	}
}
