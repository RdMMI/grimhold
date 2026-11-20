package game

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
)

var rng *rand.Rand

func SetRand(r *rand.Rand) {
	rng = r
}

func randIntN(n int) int {
	if rng != nil {
		return rng.IntN(n)
	}
	return rand.IntN(n)
}

func RollD20() int {
	return randIntN(20) + 1
}

func RollDice(diceStr string) int {
	diceStr = strings.ReplaceAll(diceStr, " ", "")
	if diceStr == "" {
		return 1
	}

	bonus := 0
	if idx := strings.IndexAny(diceStr, "+-"); idx != -1 {
		b, err := strconv.Atoi(diceStr[idx:])
		if err == nil {
			bonus = b
		}
		diceStr = diceStr[:idx]
	}

	parts := strings.Split(strings.ToLower(diceStr), "d")
	if len(parts) != 2 {
		val, err := strconv.Atoi(diceStr)
		if err != nil || val < 1 {
			return 1
		}
		return val + bonus
	}

	count, err1 := strconv.Atoi(parts[0])
	sides, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || count < 1 || sides < 1 {
		return 1
	}

	total := bonus
	for range count {
		total += randIntN(sides) + 1
	}
	if total < 1 {
		return 1
	}
	return total
}

func XPForCR(cr float64) int {
	if cr <= 0.25 {
		return 25
	}
	if cr <= 0.5 {
		return 50
	}
	return int(cr * 100)
}

func (s *State) ResolveFightRound() (string, bool, error) {
	if s.Encounter == nil {
		return "", false, fmt.Errorf("no active encounter to fight")
	}

	var log strings.Builder
	foe := &s.Encounter.Monster

	d20 := RollD20()
	attackRoll := d20 + s.Player.AttackBonus
	if d20 == 20 || attackRoll >= foe.ArmorClass {
		dmg := RollDice(s.Player.Weapon.DamageDice)
		if d20 == 20 {
			dmg *= 2
		}
		s.Encounter.CurrentHP -= dmg
		fmt.Fprintf(&log, "You strike the %s with your %s (roll %d+%d=%d vs AC %d) for %d damage!\n",
			foe.Name, s.Player.Weapon.Name, d20, s.Player.AttackBonus, attackRoll, foe.ArmorClass, dmg)
	} else {
		fmt.Fprintf(&log, "Your swing misses the %s (roll %d+%d=%d vs AC %d).\n",
			foe.Name, d20, s.Player.AttackBonus, attackRoll, foe.ArmorClass)
	}

	if s.Encounter.CurrentHP <= 0 {
		xp := XPForCR(foe.CrFloat)
		gold := randIntN(10) + 1
		s.Player.XP += xp
		s.Player.Gold += gold
		s.SlainBestiary[foe.Name]++
		s.Encounter = nil

		fmt.Fprintf(&log, "The %s collapses! You gain %d XP and %d gold.\n", foe.Name, xp, gold)
		if s.Player.CheckLevelUp() {
			fmt.Fprintf(&log, "LEVEL UP! You are now level %d (Max HP: %d, Attack Bonus: +%d)!\n",
				s.Player.Level, s.Player.MaxHP, s.Player.AttackBonus)
		}
		return log.String(), true, nil
	}

	log.WriteString(s.foeAttack())
	if s.Player.HP <= 0 {
		s.handleDefeat(&log)
		return log.String(), true, nil
	}

	fmt.Fprintf(&log, "[Your HP: %d/%d | %s HP: %d/%d]",
		s.Player.HP, s.Player.MaxHP, foe.Name, s.Encounter.CurrentHP, foe.HitPoints)
	return log.String(), false, nil
}

func (s *State) foeAttack() string {
	foe := &s.Encounter.Monster
	atkBonus := 3
	dmgDice := "1d6+1"
	atkName := "Claws"

	if len(foe.Actions) > 0 {
		act := foe.Actions[0]
		atkName = act.Name
		if act.AttackBonus > 0 {
			atkBonus = act.AttackBonus
		}
		if act.DamageDice != "" {
			dmgDice = act.DamageDice
		}
	}

	playerAC := 12 + (s.Player.Level / 2)
	d20 := RollD20()
	roll := d20 + atkBonus
	if d20 == 20 || roll >= playerAC {
		dmg := RollDice(dmgDice)
		s.Player.HP -= dmg
		return fmt.Sprintf("The %s hits you with %s (roll %d vs AC %d) for %d damage!\n",
			foe.Name, atkName, roll, playerAC, dmg)
	}
	return fmt.Sprintf("The %s attacks with %s (roll %d vs AC %d) but misses!\n",
		foe.Name, atkName, roll, playerAC)
}

func (s *State) handleDefeat(log *strings.Builder) {
	s.Encounter = nil
	s.CurrentRegion = s.World.Start
	s.Player.HP = 1
	s.RestedInArea = false
	fmt.Fprintf(log, "\nYou fall in battle... Hours later, you wake up at the %s with 1 HP and your pride wounded.",
		s.World.ByID[s.CurrentRegion].Name)
}
