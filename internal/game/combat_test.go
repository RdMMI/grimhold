package game

import (
	"math/rand/v2"
	"testing"

	"github.com/RdMMI/grimhold/internal/open5e"
	"github.com/RdMMI/grimhold/internal/world"
)

func TestXPForCR(t *testing.T) {
	tests := []struct {
		cr   float64
		want int
	}{
		{0.125, 25},
		{0.25, 25},
		{0.5, 50},
		{2.0, 200},
	}

	for _, tc := range tests {
		if got := XPForCR(tc.cr); got != tc.want {
			t.Errorf("XPForCR(%v) = %d, want %d", tc.cr, got, tc.want)
		}
	}
}

func TestCombatDeterministicVictory(t *testing.T) {
	SetRand(rand.New(rand.NewPCG(42, 99)))
	defer SetRand(nil)

	w := &world.World{
		Start: "crossroads",
		ByID: map[string]world.Region{
			"crossroads": {ID: "crossroads", Name: "Crossroads"},
		},
	}
	state := NewState(w, nil)
	state.Encounter = &ActiveEncounter{
		Monster: open5e.Monster{
			Name:       "Goblin",
			ArmorClass: 10,
			HitPoints:  5,
			CrFloat:    0.25,
		},
		CurrentHP: 5,
	}

	ended := false
	for range 10 {
		_, done, err := state.ResolveFightRound()
		if err != nil {
			t.Fatalf("unexpected combat error: %v", err)
		}
		if done {
			ended = true
			break
		}
	}

	if !ended {
		t.Fatalf("expected combat to finish within 10 rounds")
	}
	if state.SlainBestiary["Goblin"] != 1 {
		t.Fatalf("expected 1 slain Goblin in bestiary, got %d", state.SlainBestiary["Goblin"])
	}
	if state.Player.XP < 25 {
		t.Fatalf("expected at least 25 XP gained, got %d", state.Player.XP)
	}
}
