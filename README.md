# Grimhold

**Grimhold** is a text-based dungeon crawler that runs in your terminal, written in Go using only the standard library. Explore cursed regions loaded from `world.json`, encounter monsters fetched live from the [Open5e API](https://open5e.com) (D&D 5e SRD), fight in dice-based combat, level up your character, and record your kills in your bestiary.

## Prerequisites

- **Go 1.22+** installed on your machine.
- An active internet connection to fetch monster data from `https://api.open5e.com/v1/monsters/` (responses are cached in memory with a 5-minute TTL).

## Project Structure

```text
grimhold/
├── cmd/grimhold/main.go        # Thin CLI entrypoint & flags
├── internal/
│   ├── game/                   # Character, game state, dice combat, save/load & tests
│   ├── open5e/                 # Open5e HTTP client, concurrent TTL cache & tests
│   ├── repl/                   # Interactive terminal loop, command registry & ANSI colors
│   └── world/                  # World map loader & region navigation
├── world.json                  # Region definitions and connections
├── go.mod
└── README.md
```

## Build & Quality Checks

Before running or grading, you can verify formatting, static analysis, and run the network-free test suite:

```bash
go fmt ./...
go vet ./...
go test -v ./...
```

### Build the binary

**Linux / macOS:**
```bash
go build -o grimhold ./cmd/grimhold
```

**Windows (PowerShell):**
```powershell
go build -o grimhold.exe ./cmd/grimhold
```

## Running the Game

You can either run directly with `go run` or execute the compiled binary:

```bash
# Run directly
go run ./cmd/grimhold

# Or run the compiled binary (Linux/macOS)
./grimhold

# Or on Windows
.\grimhold.exe
```

### CLI Flags

- `--continue`: Automatically loads an existing `save.json` on startup.
- `--world <path>`: Specify a custom path to the world file (defaults to `world.json`).

Example:
```bash
go run ./cmd/grimhold --continue
```

## Saving & Loading

- Type `save` in the REPL at any time to persist your character stats, current region, gold, rest state, and slain bestiary to `save.json`.
- Type `load` in the REPL — or start the game with `--continue` — to restore your saved state.
- If no `save.json` exists yet, the game handles it gracefully without crashing.

## Commands Overview

| Command | Description |
|---|---|
| `help` | Dynamically list all available commands |
| `map` | Display all known regions, danger levels, and exits |
| `look` | Describe your current surroundings |
| `travel <region>` | Move to a connected region |
| `explore` | Wander the current region and trigger an encounter |
| `fight` | Engage the active foe in dice-based combat |
| `flee` | Attempt to escape combat (failure risks a free hit) |
| `character` | View your character sheet and slain bestiary |
| `rest` | Fully restore HP (once per region visit) |
| `bestiary [next\|prev]` | Browse Open5e monsters page by page |
| `inspect <monster>` | Display a monster's full stat block |
| `save` / `load` | Save or restore progress via `save.json` |
| `quit` | Exit Grimhold |

## Bonus Features Implemented (+12 pts)

1. **Test Suite (+6 pts)**: Network-free unit tests runnable via `go test -v ./...`:
   - `internal/open5e/cache_test.go`: tests TTL cache expiration using an injected fake clock.
   - `internal/open5e/client_test.go`: tests Open5e JSON decoding and cache hit behavior against an `httptest.Server`.
   - `internal/game/combat_test.go`: table-driven tests for XP scaling and deterministic combat resolution using a seeded `math/rand/v2` generator.
2. **Faster Dungeon (+4 pts)**: Traveling to a new region triggers `Client.PrefetchMonsters`, which concurrently fetches and caches all foes in that region using a `sync.WaitGroup`.
3. **ANSI Colors (+2 pts)**: Hand-rolled ANSI escape codes render color-coded HP bars (`character`) and region danger badges (`map`, `look`).

## Sample Session

```text
$ go run ./cmd/grimhold
Welcome to Grimhold. Type 'help' for a list of commands.
=== The Weeping Crossroads (crossroads) ===
A crooked gallows stands where three muddy paths meet.
Danger 1 (Low)
Exits: ashen-marsh, sunken-crypt

grimhold > map
[*] crossroads       (Danger 1 (Low)) -> Exits: ashen-marsh, sunken-crypt
[ ] ashen-marsh      (Danger 2 (Moderate)) -> Exits: crossroads, drowned-chapel
[ ] sunken-crypt     (Danger 2 (Moderate)) -> Exits: crossroads, citadel-of-ash
[ ] drowned-chapel   (Danger 3 (Deadly)) -> Exits: ashen-marsh, citadel-of-ash
[ ] citadel-of-ash   (Danger 4 (Deadly)) -> Exits: sunken-crypt, drowned-chapel

grimhold > travel ashen-marsh
You trudge into Ashen Marsh. Fog, iron, and things that breathe under the water.

grimhold > explore
A Skeleton emerges from the shadows! (HP: 13, AC: 13, CR: 1/4)
Type 'fight' to attack or 'flee' to run!

grimhold > fight
You strike the Skeleton with your Rusty Longsword (roll 14+4=18 vs AC 13) for 8 damage!
The Skeleton attacks with Shortsword (roll 10 vs AC 12) but misses!
[Your HP: 25/25 | Skeleton HP: 5/13]
You strike the Skeleton with your Rusty Longsword (roll 17+4=21 vs AC 13) for 6 damage!
The Skeleton collapses! You gain 25 XP and 7 gold.

grimhold > character
=== Wanderer (Level 1) ===
[##########] 25/25 HP | XP: 25/50 | Gold: 7
Weapon: Rusty Longsword (1d8+2) | Attack Bonus: +4

Slain Bestiary:
  - Skeleton: 1

grimhold > inspect goblin
=== Goblin (Small humanoid, neutral evil) ===
AC: 15 | HP: 7 (2d6) | CR: 1/4
Actions:
  * Scimitar: Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 5 (1d6 + 2) slashing damage.
  * Shortbow: Ranged Weapon Attack: +4 to hit, range 80/320 ft., one target. Hit: 5 (1d6 + 2) piercing damage.

grimhold > save
Game saved to save.json

grimhold > quit
Farewell, wanderer.
```