package repl

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"os"
	"sort"
	"strings"

	"github.com/RdMMI/grimhold/internal/game"
)

type cliCommand struct {
	name        string
	description string
	callback    func(s *game.State, args []string) error
}

func getCommands() map[string]cliCommand {
	var cmds map[string]cliCommand
	cmds = map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Show this message",
			callback: func(s *game.State, args []string) error {
				fmt.Println("Grimhold — a dungeon crawler\n\nCommands:")
				keys := make([]string, 0, len(cmds))
				for k := range cmds {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					fmt.Printf("  %-22s %s\n", cmds[k].name, cmds[k].description)
				}
				return nil
			},
		},
		"map": {
			name:        "map",
			description: "List the known regions",
			callback:    cmdMap,
		},
		"look": {
			name:        "look",
			description: "Describe your surroundings",
			callback:    cmdLook,
		},
		"travel": {
			name:        "travel <region>",
			description: "Travel to a connected region",
			callback:    cmdTravel,
		},
		"explore": {
			name:        "explore",
			description: "Wander the region (risk an encounter)",
			callback:    cmdExplore,
		},
		"fight": {
			name:        "fight",
			description: "Fight the current foe",
			callback:    cmdFight,
		},
		"flee": {
			name:        "flee",
			description: "Attempt to flee from combat",
			callback:    cmdFlee,
		},
		"character": {
			name:        "character",
			description: "Print your character sheet and slain bestiary",
			callback:    cmdCharacter,
		},
		"rest": {
			name:        "rest",
			description: "Restore HP (once per region visit)",
			callback:    cmdRest,
		},
		"bestiary": {
			name:        "bestiary [next|prev]",
			description: "Browse Open5e monsters page by page",
			callback:    cmdBestiary,
		},
		"inspect": {
			name:        "inspect <monster>",
			description: "Inspect a monster's stat block",
			callback:    cmdInspect,
		},
		"save": {
			name:        "save",
			description: "Save game to save.json",
			callback:    cmdSave,
		},
		"load": {
			name:        "load",
			description: "Load game from save.json",
			callback:    cmdLoad,
		},
		"quit": {
			name:        "quit",
			description: "Exit the game",
			callback: func(s *game.State, args []string) error {
				fmt.Println("Farewell, wanderer.")
				os.Exit(0)
				return nil
			},
		},
	}
	return cmds
}

func Start(s *game.State) {
	scanner := bufio.NewScanner(os.Stdin)
	commands := getCommands()

	fmt.Println("Welcome to Grimhold. Type 'help' for a list of commands.")
	_ = cmdLook(s, nil)

	for {
		fmt.Print("grimhold > ")
		if !scanner.Scan() {
			break
		}

		words := strings.Fields(strings.TrimSpace(scanner.Text()))
		if len(words) == 0 {
			continue
		}

		cmdName := strings.ToLower(words[0])
		args := words[1:]

		cmd, exists := commands[cmdName]
		if !exists {
			fmt.Printf("Unknown command %q. Type 'help' for commands.\n", cmdName)
			continue
		}

		if err := cmd.callback(s, args); err != nil {
			fmt.Println("Error:", err)
		}
	}
}

func cmdMap(s *game.State, args []string) error {
	for _, r := range s.World.Regions {
		marker := " "
		if r.ID == s.CurrentRegion {
			marker = colorCyan + "*" + colorReset
		}
		fmt.Printf("[%s] %-16s (%s) -> Exits: %s\n",
			marker, r.ID, dangerBadge(r.Danger), strings.Join(r.Exits, ", "))
	}
	return nil
}

func cmdLook(s *game.State, args []string) error {
	r := s.World.ByID[s.CurrentRegion]
	fmt.Printf("%s=== %s (%s) ===%s\n%s\n%s\nExits: %s\n",
		colorBold, r.Name, r.ID, colorReset, r.Description, dangerBadge(r.Danger), strings.Join(r.Exits, ", "))
	return nil
}

func cmdTravel(s *game.State, args []string) error {
	if s.Encounter != nil {
		return fmt.Errorf("you cannot travel while in combat! Use 'fight' or 'flee'")
	}
	if len(args) < 1 {
		return fmt.Errorf("usage: travel <region-id>")
	}
	target := strings.ToLower(args[0])
	if !s.World.CanTravel(s.CurrentRegion, target) {
		return fmt.Errorf("cannot travel to %q from here", target)
	}
	s.CurrentRegion = target
	s.RestedInArea = false
	r := s.World.ByID[target]

	go s.Client.PrefetchMonsters(r.Foes)

	fmt.Printf("You trudge into %s%s%s. %s\n", colorCyan, r.Name, colorReset, r.Description)
	return nil
}

func cmdExplore(s *game.State, args []string) error {
	if s.Encounter != nil {
		return fmt.Errorf("you are already facing a %s! Use 'fight' or 'flee'", s.Encounter.Monster.Name)
	}
	region := s.World.ByID[s.CurrentRegion]
	if len(region.Foes) == 0 {
		fmt.Println("The area is eerily quiet.")
		return nil
	}

	var worst *game.ActiveEncounter
	rolls := max(1, region.Danger)
	for range rolls {
		slug := region.Foes[rand.IntN(len(region.Foes))]
		m, err := s.Client.GetMonster(slug)
		if err != nil {
			return err
		}
		if worst == nil || m.CrFloat > worst.Monster.CrFloat {
			worst = &game.ActiveEncounter{Monster: *m, CurrentHP: m.HitPoints}
		}
	}

	s.Encounter = worst
	fmt.Printf("A %s emerges from the shadows! (HP: %d, AC: %d, CR: %s)\nType 'fight' to attack or 'flee' to run!\n",
		worst.Monster.Name, worst.Monster.HitPoints, worst.Monster.ArmorClass, worst.Monster.ChallengeRating)
	return nil
}

func cmdFight(s *game.State, args []string) error {
	for s.Encounter != nil {
		msg, ended, err := s.ResolveFightRound()
		if err != nil {
			return err
		}
		fmt.Println(msg)
		if ended {
			break
		}
	}
	return nil
}

func cmdFlee(s *game.State, args []string) error {
	if s.Encounter == nil {
		return fmt.Errorf("there is nothing to flee from")
	}
	roll := game.RollD20()
	if roll >= 10 {
		fmt.Printf("You roll %d and slip away from the %s!\n", roll, s.Encounter.Monster.Name)
		s.Encounter = nil
		return nil
	}
	fmt.Printf("You fail to flee (rolled %d)! The %s gets a free strike!\n", roll, s.Encounter.Monster.Name)
	dmg := game.RollDice("1d6+1")
	s.Player.HP -= dmg
	fmt.Printf("You take %d damage (HP: %d/%d).\n", dmg, s.Player.HP, s.Player.MaxHP)
	if s.Player.HP <= 0 {
		s.Encounter = nil
		s.CurrentRegion = s.World.Start
		s.Player.HP = 1
		fmt.Println("You collapse and wake up back at the crossroads with 1 HP.")
	}
	return nil
}

func cmdCharacter(s *game.State, args []string) error {
	p := s.Player
	fmt.Printf("%s=== %s (Level %d) ===%s\n%s | XP: %d/%d | Gold: %d\nWeapon: %s (%s) | Attack Bonus: +%d\n",
		colorBold, p.Name, p.Level, colorReset, hpBar(p.HP, p.MaxHP), p.XP, p.Level*p.Level*50, p.Gold, p.Weapon.Name, p.Weapon.DamageDice, p.AttackBonus)
	if len(s.SlainBestiary) > 0 {
		fmt.Println("\nSlain Bestiary:")
		for name, count := range s.SlainBestiary {
			fmt.Printf("  - %s: %d\n", name, count)
		}
	}
	return nil
}

func cmdRest(s *game.State, args []string) error {
	if s.Encounter != nil {
		return fmt.Errorf("cannot rest during combat")
	}
	if s.RestedInArea {
		return fmt.Errorf("you have already rested in this region; travel elsewhere first")
	}
	s.Player.HP = s.Player.MaxHP
	s.RestedInArea = true
	fmt.Printf("You make a small camp and tend your wounds. HP restored to %d/%d.\n", s.Player.HP, s.Player.MaxHP)
	return nil
}

func cmdBestiary(s *game.State, args []string) error {
	url := ""
	if len(args) > 0 {
		switch strings.ToLower(args[0]) {
		case "next":
			if s.NextPageURL == "" {
				return fmt.Errorf("no next page available")
			}
			url = s.NextPageURL
		case "prev":
			if s.PrevPageURL == "" {
				return fmt.Errorf("no previous page available")
			}
			url = s.PrevPageURL
		}
	}
	list, err := s.Client.ListMonsters(url)
	if err != nil {
		return err
	}
	s.NextPageURL, s.PrevPageURL = "", ""
	if list.Next != nil {
		s.NextPageURL = *list.Next
	}
	if list.Previous != nil {
		s.PrevPageURL = *list.Previous
	}
	for _, m := range list.Results {
		fmt.Printf("- %-25s (slug: %-20s CR: %s, HP: %d)\n", m.Name, m.Slug, m.ChallengeRating, m.HitPoints)
	}
	return nil
}

func cmdInspect(s *game.State, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: inspect <monster-slug>")
	}
	m, err := s.Client.GetMonster(strings.ToLower(args[0]))
	if err != nil {
		return err
	}
	fmt.Printf("=== %s (%s %s, %s) ===\nAC: %d | HP: %d (%s) | CR: %s\nActions:\n",
		m.Name, m.Size, m.Type, m.Alignment, m.ArmorClass, m.HitPoints, m.HitDice, m.ChallengeRating)
	for _, a := range m.Actions {
		fmt.Printf("  * %s: %s\n", a.Name, a.Desc)
	}
	return nil
}

func cmdSave(s *game.State, args []string) error {
	if err := s.Save(game.DefaultSaveFile); err != nil {
		return err
	}
	fmt.Println("Game saved to", game.DefaultSaveFile)
	return nil
}

func cmdLoad(s *game.State, args []string) error {
	if err := s.Load(game.DefaultSaveFile); err != nil {
		return fmt.Errorf("could not load save (does %s exist?): %w", game.DefaultSaveFile, err)
	}
	fmt.Println("Save loaded successfully.")
	return cmdLook(s, nil)
}
