package game

import (
	"github.com/RdMMI/grimhold/internal/open5e"
	"github.com/RdMMI/grimhold/internal/world"
)

type Weapon struct {
	Name       string `json:"name"`
	DamageDice string `json:"damage_dice"`
}

type Character struct {
	Name        string `json:"name"`
	Level       int    `json:"level"`
	XP          int    `json:"xp"`
	HP          int    `json:"hp"`
	MaxHP       int    `json:"max_hp"`
	AttackBonus int    `json:"attack_bonus"`
	Gold        int    `json:"gold"`
	Weapon      Weapon `json:"weapon"`
}

type ActiveEncounter struct {
	Monster   open5e.Monster
	CurrentHP int
}

type State struct {
	Player        Character        `json:"player"`
	CurrentRegion string           `json:"current_region"`
	RestedInArea  bool             `json:"rested_in_area"`
	SlainBestiary map[string]int   `json:"slain_bestiary"`
	NextPageURL   string           `json:"-"`
	PrevPageURL   string           `json:"-"`
	Encounter     *ActiveEncounter `json:"-"`
	World         *world.World     `json:"-"`
	Client        *open5e.Client   `json:"-"`
}

func NewState(w *world.World, client *open5e.Client) *State {
	return &State{
		Player: Character{
			Name:        "Wanderer",
			Level:       1,
			XP:          0,
			HP:          25,
			MaxHP:       25,
			AttackBonus: 4,
			Gold:        0,
			Weapon: Weapon{
				Name:       "Rusty Longsword",
				DamageDice: "1d8+2",
			},
		},
		CurrentRegion: w.Start,
		RestedInArea:  false,
		SlainBestiary: make(map[string]int),
		World:         w,
		Client:        client,
	}
}

func (c *Character) CheckLevelUp() bool {
	leveled := false
	for c.XP >= c.Level*c.Level*50 {
		c.Level++
		c.MaxHP += 8
		c.HP = c.MaxHP
		c.AttackBonus += 1
		leveled = true
	}
	return leveled
}
