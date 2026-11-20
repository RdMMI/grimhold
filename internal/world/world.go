package world

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
)

type Region struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Danger      int      `json:"danger"`
	Foes        []string `json:"foes"`
	Exits       []string `json:"exits"`
}

type World struct {
	Start   string            `json:"start"`
	Regions []Region          `json:"regions"`
	ByID    map[string]Region `json:"-"`
}

func Load(path string) (*World, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading world file: %w", err)
	}

	var w World
	if err := json.Unmarshal(data, &w); err != nil {
		return nil, fmt.Errorf("parsing world JSON: %w", err)
	}

	w.ByID = make(map[string]Region, len(w.Regions))
	for _, r := range w.Regions {
		w.ByID[r.ID] = r
	}

	if _, ok := w.ByID[w.Start]; !ok {
		return nil, fmt.Errorf("start region %q not found in regions", w.Start)
	}

	return &w, nil
}

func (w *World) CanTravel(currentID, targetID string) bool {
	curr, ok := w.ByID[currentID]
	if !ok {
		return false
	}
	return slices.Contains(curr.Exits, targetID)
}
