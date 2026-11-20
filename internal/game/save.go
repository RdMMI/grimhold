package game

import (
	"encoding/json"
	"fmt"
	"os"
)

const DefaultSaveFile = "save.json"

func (s *State) Save(filename string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling save state: %w", err)
	}
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("writing save file: %w", err)
	}
	return nil
}

func (s *State) Load(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("reading save file: %w", err)
	}

	var loaded State
	if err := json.Unmarshal(data, &loaded); err != nil {
		return fmt.Errorf("unmarshaling save file: %w", err)
	}

	s.Player = loaded.Player
	s.CurrentRegion = loaded.CurrentRegion
	s.RestedInArea = loaded.RestedInArea
	if loaded.SlainBestiary != nil {
		s.SlainBestiary = loaded.SlainBestiary
	} else {
		s.SlainBestiary = make(map[string]int)
	}
	s.Encounter = nil
	return nil
}
