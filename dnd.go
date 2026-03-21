package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// ==================== CONSTANTS (easily adjustable) ====================

const (
	XPPerLevel        = 5     // XP needed to level up (original: 10)
	HealerSingleAmt   = 4     // healer healing one target
	HealerSingleAmtOP = 8     // healer OP heal one target
	HealerAllAmt      = 1     // healer healing all
	HealerAllAmtOP    = 2     // healer OP heal all
	PlayerHealAmt     = 2     // non-healer healing a teammate
	PlayerHealAmtOP   = 3     // non-healer OP heal
	HealDMApplied     = false // set true to apply enemy DM to heal rolls
	UseD12            = false // false = use 2d6 for all checks
)

// ==================== TYPES ====================

type Weapon struct {
	Name       string
	Damage     int
	AttackStat string // "str", "agi", "int"
}

type Character struct {
	Name         string
	Class        string
	Race         string
	Strength     int
	Agility      int
	Intelligence int
	Charisma     int
	HP           int
	MaxHP        int
	XP           int
	Level        int
	Weapon       Weapon
	IsHealer     bool
	Alive        bool
}

type Enemy struct {
	Name               string
	HP                 int
	MaxHP              int
	Damage             int
	DifficultyModifier int // stored as negative (e.g. -3)
	Alive              bool
}

// ==================== DICE ====================

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func rollD12() int { return rng.Intn(12) + 1 }
func rollD6() int  { return rng.Intn(6) + 1 }

func roll() (int, bool, bool) {
	if UseD12 {
		r := rollD12()
		return r, r == 1, r == 12
	}
	r := rollD6() + rollD6()
	return r, r == 2, r == 12
}

func check(stat, dm int) (bool, bool, int) {
	target := stat + dm
	fmt.Printf("  To Roll: ≤%v\n", target)
	if target < 1 {
		target = 1
	}
	r, opS, opF := roll()
	fmt.Printf("  Rolled:   %v\n", r)
	if opF {
		return false, false, r
	}
	if opS {
		return true, true, r
	}
	return r <= target, false, r
}

// ==================== DATA TABLES ====================

var classBase = map[string][4]int{
	"fighter":       {6, 5, 5, 4},
	"mage":          {4, 5, 6, 5},
	"rogue":         {4, 6, 5, 5},
	"warrior-mage":  {5, 4, 6, 5},
	"mage-rogue":    {4, 5, 6, 5},
	"rogue-fighter": {5, 5, 4, 6},
}

var classBonus = map[string][4]int{
	"fighter":       {1, 0, 0, 0},
	"mage":          {0, 0, 1, 0},
	"rogue":         {0, 1, 0, 0},
	"warrior-mage":  {1, 0, 0, 0},
	"mage-rogue":    {0, 0, 1, 0},
	"rogue-fighter": {0, 1, 0, 0},
}

var raceBonus = map[string][4]int{
	"orc":       {1, 0, 0, 0},
	"troll":     {1, 0, 1, 0},
	"night elf": {0, 0, 1, 0},
	"undead":    {0, 1, 1, 0},
	"harpy":     {0, 1, 0, 0},
	"goliath":   {1, 1, 0, 0},
}

var classWeapon = map[string]Weapon{
	"fighter":       {"Two-handed Axe", 7, "str"},
	"mage":          {"Staff", 5, "int"},
	"rogue":         {"Daggers", 6, "agi"},
	"warrior-mage":  {"Sword", 6, "str"},
	"mage-rogue":    {"Arcane Blade", 6, "int"},
	"rogue-fighter": {"Short Sword", 6, "agi"},
	"mage-healer":   {"Healing Staff", 3, "int"},
}

var enemyPresets = map[string]Enemy{
	"goblin":      {"Goblin", 12, 12, 5, -3, true},
	"vine blight": {"Vine Blight", 18, 18, 6, -1, true},
	"bugbear":     {"Bugbear", 45, 45, 10, -5, true},
	"wolf":        {"Wolf", 20, 20, 6, -1, true},
	"orc":         {"Orc", 30, 30, 8, -4, true},
	"karguk":      {"Karguk (Boss)", 45, 45, 11, -5, true},
}

var defaultParty = []struct {
	class, race string
	healer      bool
}{
	{"mage", "night elf", true},
	{"fighter", "orc", false},
	{"fighter", "orc", false},
	{"rogue", "harpy", false},
}

// ==================== BEST-FIT SUGGESTIONS ====================

var bestRace = map[string]string{
	"fighter":       "orc",
	"warrior-mage":  "troll",
	"mage":          "night elf",
	"mage-rogue":    "undead",
	"rogue":         "harpy",
	"rogue-fighter": "goliath",
}

func bestRaceForClass(class string) string {
	if r, ok := bestRace[class]; ok {
		return r
	}
	return "orc"
}

// ==================== CHARACTER ====================

func newCharacter(name, class, race string, isHealer bool) *Character {
	base := classBase[class]
	cb := classBonus[class]
	rb := raceBonus[race]

	str := base[0] + cb[0] + rb[0]
	agi := base[1] + cb[1] + rb[1]
	intel := base[2] + cb[2] + rb[2]
	cha := base[3] + cb[3] + rb[3]
	hp := 5 + agi + str

	return &Character{
		Name: name, Class: class, Race: race,
		Strength: str, Agility: agi, Intelligence: intel, Charisma: cha,
		HP: hp, MaxHP: hp,
		Level: 1, Weapon: classWeapon[class],
		IsHealer: isHealer, Alive: true,
	}
}

func (c *Character) attackStat() int {
	switch c.Weapon.AttackStat {
	case "agi":
		return c.Agility
	case "int":
		return c.Intelligence
	default:
		return c.Strength
	}
}

func (c *Character) damage() int {
	return c.attackStat() + c.Weapon.Damage
}

// ==================== COMBAT ACTIONS ====================

func (c *Character) doAttack(e *Enemy, reader *bufio.Reader) {
	fmt.Printf("Attacking %v\n", e.Name)
	fmt.Printf("  %v check:\n", c.Weapon.AttackStat)
	success, op, _ := check(c.attackStat(), e.DifficultyModifier)
	if op {
		c.gainXP(reader)
	}
	if success {
		dmg := c.damage()
		e.HP -= dmg
		if e.HP < 0 {
			e.HP = 0
		}
		tag := opTag(op)
		fmt.Printf("  %s hits %s for %d dmg%s  [%s HP: %d/%d]\n",
			c.Name, e.Name, dmg, tag, e.Name, e.HP, e.MaxHP)
		if e.HP == 0 {
			e.Alive = false
			fmt.Printf("  💀 %s defeated!\n", e.Name)
		}
	} else {
		fmt.Printf("  %s misses %s.\n", c.Name, e.Name)
	}
}

func (e *Enemy) doAttack(party []*Character, reader *bufio.Reader) {
	alive := alivePlayers(party)
	if len(alive) == 0 {
		return
	}
	target := alive[rng.Intn(len(alive))]
	fmt.Printf("  Attacking %v\n", target.Name)
	// Dodge check: player Agility + enemy DM (DM is negative)
	success, op, _ := check(target.Agility, e.DifficultyModifier)
	if op {
		target.gainXP(reader)
	}
	if success {
		tag := ""
		if op {
			tag = " overwhelmingly"
		}
		fmt.Printf("  %s%s dodges %s's attack!\n", target.Name, tag, e.Name)
	} else {
		target.HP -= e.Damage
		if target.HP < 0 {
			target.HP = 0
		}
		fmt.Printf("  %s hits %s for %d dmg  [%s HP: %d/%d]\n",
			e.Name, target.Name, e.Damage, target.Name, target.HP, target.MaxHP)
		if target.HP == 0 {
			target.Alive = false
			fmt.Printf("  💀 %s is knocked out!\n", target.Name)
		}
	}
}

// ==================== HEALING ====================

func healDecision(healer *Character, party []*Character) string {
	crit, low := 0, 0
	for _, p := range party {
		if !p.Alive || p == healer {
			continue
		}
		ratio := float64(p.HP) / float64(p.MaxHP)
		if ratio < 0.50 { // was 0.3 default
			crit++
		} else if ratio < 0.90 { // was 0.5 default
			low++
		}
	}
	if crit+low >= 2 {
		return "heal_all"
	}
	if crit >= 1 || low >= 1 {
		return "heal_single"
	}
	return "attack"
}

func mostCritical(healer *Character, party []*Character) *Character {
	var target *Character
	minRatio := 2.0
	for _, p := range party {
		if !p.Alive || p == healer {
			continue
		}
		r := float64(p.HP) / float64(p.MaxHP)
		if r < minRatio {
			minRatio = r
			target = p
		}
	}
	return target
}

func (c *Character) doHeal(party []*Character, reader *bufio.Reader) {
	decision := healDecision(c, party)
	if decision == "heal_all" {
		fmt.Printf("Healing the party\n")
	} else {
		target := mostCritical(c, party)
		if target != nil {
			fmt.Printf("Healing %s\n", target.Name)
		}
	}
	fmt.Printf("  int check:\n")
	success, op, _ := check(c.Intelligence, 0) // HealDMApplied: change 0 to dm if needed
	if op {
		c.gainXP(reader)
	}
	if !success {
		fmt.Printf("  %s loses concentration — heal fails.\n", c.Name)
		return
	}

	if decision == "heal_all" {
		amt := HealerAllAmt
		if op {
			amt = HealerAllAmtOP
		}
		fmt.Printf("  %s heals the whole party for %d HP%s!\n", c.Name, amt, opTag(op))
		for _, p := range party {
			if p.Alive {
				p.HP = min(p.HP+amt, p.MaxHP)
			}
		}
	} else {
		target := mostCritical(c, party)
		if target == nil {
			return
		}
		amt := HealerSingleAmt
		if op {
			amt = HealerSingleAmtOP
		}
		target.HP = min(target.HP+amt, target.MaxHP)
		fmt.Printf("  %s heals %s for %d HP%s  [%s HP: %d/%d]\n",
			c.Name, target.Name, amt, opTag(op), target.Name, target.HP, target.MaxHP)
	}
}

// ==================== XP & LEVELING ====================

func (c *Character) gainXP(reader *bufio.Reader) {
	c.XP++
	fmt.Printf("  ✦ %s gains XP! (%d/%d)\n", c.Name, c.XP, XPPerLevel)
	if c.XP >= XPPerLevel {
		c.XP = 0
		c.Level++
		fmt.Printf("\n  🎉 %s reached Level %d! Choose a stat to upgrade:\n", c.Name, c.Level)
		c.chooseStat(reader)
	}
}

func (c *Character) chooseStat(reader *bufio.Reader) {
	fmt.Printf("\n  STR:%d AGI:%d INT:%d CHA:%d\n",
		c.Strength, c.Agility, c.Intelligence, c.Charisma)
	fmt.Println("  1. Strength  2. Agility  3. Intelligence  4. Charisma")
	for {
		fmt.Print("  > ")
		input, _ := reader.ReadString('\n')
		switch strings.TrimSpace(input) {
		case "1":
			c.Strength++
			fmt.Printf("  ▲ Strength → %d\n", c.Strength)
			return
		case "2":
			c.Agility++
			c.MaxHP++
			c.HP++
			fmt.Printf("  ▲ Agility → %d (MaxHP now %d)\n", c.Agility, c.MaxHP)
			return
		case "3":
			c.Intelligence++
			c.MaxHP++
			c.HP++
			fmt.Printf("  ▲ Intelligence → %d (MaxHP now %d)\n", c.Intelligence, c.MaxHP)
			return
		case "4":
			c.Charisma++
			fmt.Printf("  ▲ Charisma → %d\n", c.Charisma)
			return
		default:
			fmt.Println("  Enter 1, 2, 3, or 4.")
		}
	}
}

func manualUpgrade(party []*Character, reader *bufio.Reader) {
	fmt.Println("\nWhich player to upgrade?")
	for i, p := range party {
		if p.Alive {
			fmt.Printf("  %d. %s (STR:%d AGI:%d INT:%d CHA:%d)\n",
				i+1, p.Name, p.Strength, p.Agility, p.Intelligence, p.Charisma)
		}
	}
	fmt.Print("  > ")
	input, _ := reader.ReadString('\n')
	idx, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || idx < 1 || idx > len(party) {
		fmt.Println("  Invalid selection.")
		return
	}
	p := party[idx-1]
	if !p.Alive {
		fmt.Println("  That character is knocked out.")
		return
	}
	p.chooseStat(reader)
}

// ==================== HELPERS ====================

func alivePlayers(party []*Character) []*Character {
	var out []*Character
	for _, p := range party {
		if p.Alive {
			out = append(out, p)
		}
	}
	return out
}

func aliveEnemies(enemies []*Enemy) []*Enemy {
	var out []*Enemy
	for _, e := range enemies {
		if e.Alive {
			out = append(out, e)
		}
	}
	return out
}

func lowestHPEnemy(enemies []*Enemy) *Enemy {
	var target *Enemy
	minRatio := 2.0
	for _, e := range enemies {
		if !e.Alive {
			continue
		}
		r := float64(e.HP) / float64(e.MaxHP)
		if r < minRatio {
			minRatio = r
			target = e
		}
	}
	return target
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func opTag(op bool) string {
	if op {
		return " (★OP)"
	}
	return ""
}

func printParty(party []*Character) {
	fmt.Println("┌─ Party ──────────────────────────────────────────────────────────┐")
	for _, p := range party {
		status := "✓"
		if !p.Alive {
			status = "✗"
		}
		healer := ""
		if p.IsHealer {
			healer = " [healer]"
		}
		fmt.Printf("│ [%s] %-10s %-15s  HP:%2d/%-2d  STR:%d AGI:%d INT:%d CHA:%d  XP:%d/%d  Lvl:%d%s\n",
			status, p.Name, p.Race+" "+p.Class,
			p.HP, p.MaxHP, p.Strength, p.Agility, p.Intelligence, p.Charisma,
			p.XP, XPPerLevel, p.Level, healer)
	}
	fmt.Println("└──────────────────────────────────────────────────────────────────┘")
}

func printEnemies(enemies []*Enemy) {
	fmt.Println("┌─ Enemies ───────────────────────────────────────────┐")
	for _, e := range enemies {
		if e.Alive {
			fmt.Printf("│  %-20s  HP:%2d/%-2d  DMG:%d  DM:%d\n",
				e.Name, e.HP, e.MaxHP, e.Damage, e.DifficultyModifier)
		}
	}
	fmt.Println("└─────────────────────────────────────────────────────┘")
}

// ==================== SETUP ====================

const maxInputLen = 30

// sanitize strips control characters (ANSI escapes, etc.) to prevent
// terminal injection in a multi-user SSH environment.
func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		if !unicode.IsControl(r) {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if len(out) > maxInputLen {
		out = out[:maxInputLen]
	}
	return out
}

func readLine(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func readInt(reader *bufio.Reader, prompt string) int {
	for {
		s := readLine(reader, prompt)
		n, err := strconv.Atoi(s)
		if err == nil {
			return n
		}
		fmt.Println("  Please enter a number.")
	}
}

func setupParty(reader *bufio.Reader) []*Character {
	n := readInt(reader, "Number of players (1-6): ")
	if n < 1 {
		n = 1
	}
	if n > 6 {
		n = 6
	}

	fmt.Println("\nClasses: fighter, mage, rogue, warrior-mage, mage-rogue, rogue-fighter")
	fmt.Println("Races:   orc, troll, night elf, undead, harpy, goliath")
	fmt.Println("\n(Press Enter to use defaults)\n")

	party := make([]*Character, 0, n)
	for i := 0; i < n; i++ {
		def := struct {
			class, race string
			healer      bool
		}{"fighter", "orc", false}
		if i < len(defaultParty) {
			def = defaultParty[i]
		}

		fmt.Printf("── Player %d ──────────────────\n", i+1)
		name := sanitize(readLine(reader, fmt.Sprintf("\n  Name (default: Player%d): ", i+1)))
		if name == "" {
			name = fmt.Sprintf("Player%d", i+1)
		}

		var class string
		for {
			class = strings.ToLower(readLine(reader, fmt.Sprintf("  Class (default: %s): ", def.class)))
			if class == "" {
				class = def.class
			}
			if _, ok := classBase[class]; ok {
				break
			}
			fmt.Println("  No such class. Options: fighter, mage, rogue, warrior-mage, mage-rogue, rogue-fighter")
		}

		// Suggest the best race for the chosen class
		suggestedRace := bestRaceForClass(class)
		var race string
		for {
			race = strings.ToLower(readLine(reader, fmt.Sprintf("  Race (default: %s): ", suggestedRace)))
			if race == "" {
				race = suggestedRace
			}
			if _, ok := raceBonus[race]; ok {
				break
			}
			fmt.Println("  No such race. Options: orc, troll, night elf, undead, harpy, goliath")
		}

		isHealer := class == "mage"
		if class == "mage" {
			h := strings.ToLower(readLine(reader, "  Healing mage? (y/n, default: y): "))
			isHealer = h != "n"
		}

		c := newCharacter(name, class, race, isHealer)
		party = append(party, c)
		fmt.Printf("  → %s | HP:%d | STR:%d AGI:%d INT:%d CHA:%d | %s (%d dmg)\n\n",
			c.Name, c.MaxHP, c.Strength, c.Agility, c.Intelligence, c.Charisma,
			c.Weapon.Name, c.Weapon.Damage)
	}
	return party
}

func setupEnemies(reader *bufio.Reader) []*Enemy {
	fmt.Println("\n── Enemy Setup ──────────────────────────────────────")
	fmt.Println("Presets: goblin, vine blight, bugbear, wolf, orc, karguk")
	fmt.Println("Type 'custom' to define your own, 'done' when finished.\n")

	enemies := make([]*Enemy, 0)
	for {
		input := strings.ToLower(readLine(reader, "Add enemy: "))
		if input == "done" || input == "" {
			break
		}

		if input == "custom" {
			name := sanitize(readLine(reader, "  Name: "))
			hp := readInt(reader, "  HP: ")
			dmg := readInt(reader, "  Damage: ")
			dm := readInt(reader, "  Difficulty Modifier (e.g. -3): ")
			e := &Enemy{Name: name, HP: hp, MaxHP: hp, Damage: dmg, DifficultyModifier: dm, Alive: true}
			enemies = append(enemies, e)
			fmt.Printf("  Added %s\n", name)
		} else if preset, ok := enemyPresets[input]; ok {
			// count := readInt(reader, fmt.Sprintf("  How many?: "))
			count := readInt(reader, "  How many?: ")
			if count < 1 {
				count = 1
			}

			for j := 0; j < count; j++ {
				e := preset
				if count > 1 {
					e.Name = fmt.Sprintf("%s %d", preset.Name, j+1)
				}
				enemies = append(enemies, &e)
			}
			fmt.Printf("  Added %d × %s\n", count, preset.Name)
		} else {
			fmt.Println("  Unknown — try a preset name, 'custom', or 'done'.")
			continue
		}

		more := strings.ToLower(readLine(reader, "Add another? (y/n): "))
		if more != "y" {
			break
		}
	}

	if len(enemies) == 0 {
		fmt.Println("No enemies added — defaulting to 3 goblins.")
		for i := 1; i <= 3; i++ {
			g := enemyPresets["goblin"]
			g.Name = fmt.Sprintf("Goblin %d", i)
			enemies = append(enemies, &g)
		}
	}

	return enemies
}

// ==================== COMBAT LOOP ====================

func runCombat(party []*Character, enemies []*Enemy, reader *bufio.Reader) {
	for round := 1; ; round++ {
		ap := alivePlayers(party)
		ae := aliveEnemies(enemies)

		if len(ap) == 0 {
			fmt.Println("\n💀 All players are knocked out. Defeat.")
			break
		}
		if len(ae) == 0 {
			fmt.Println("\n🏆 All enemies defeated! Victory!")
			break
		}

		fmt.Printf("\n╔══════════════ ROUND %d ══════════════╗\n", round)
		printParty(party)
		printEnemies(enemies)

		upg := strings.ToLower(readLine(reader, "\nUpgrade a stat? (y/n, default: n): "))
		// if upg == "" {
		// 	upg = "n"
		// } else if upg == "y" {
		// 	manualUpgrade(party, reader)
		// }
		switch upg {
		case "":
			upg = "n"
		case "y":
			manualUpgrade(party, reader)
		}

		fmt.Println("\n── Player Turns ──")
		for _, p := range party {
			if !p.Alive {
				continue
			}
			ae = aliveEnemies(enemies)
			if len(ae) == 0 {
				break
			}
			fmt.Printf("\n  %s's turn: ", p.Name)

			if p.IsHealer { // If the player's a healer
				decision := healDecision(p, party) // Take a decision between attacking and healing
				if decision == "attack" {
					target := lowestHPEnemy(enemies) // If attacking, set the lowest hp enemy as the target
					if target != nil {               // attack only if there is a target
						p.doAttack(target, reader)
					}
				} else {
					p.doHeal(party, reader) // otherwise, heal
				}
			} else { // If player is not a healer
				target := lowestHPEnemy(enemies) // set the lowest hp enemy as the target
				if target != nil {
					p.doAttack(target, reader) // attack them
				}
			}
		}

		ae = aliveEnemies(enemies)
		if len(ae) > 0 {
			fmt.Println("\n── Enemy Turns ──")
			for _, e := range enemies {
				if !e.Alive {
					continue
				}
				if len(alivePlayers(party)) == 0 {
					break
				}
				fmt.Printf("\n  %s attacks!\n", e.Name)
				e.doAttack(party, reader)
			}
		}

		readLine(reader, "\nPress Enter for next round...")
	}

	fmt.Println("\n──── Final Status ────")
	printParty(party)
}

// ==================== PARTY MANAGEMENT ====================

func manageParty(party []*Character, reader *bufio.Reader) []*Character {
	for {
		fmt.Println("\n── Current Party ──")
		printParty(party)
		fmt.Println("\n  1. Keep party as-is")
		fmt.Println("  2. Remove a player")
		fmt.Println("  3. Edit a player")
		fmt.Println("  4. Add a player")
		fmt.Println("  5. Scrap and build a new party")

		choice := readLine(reader, "\n  > ")
		switch choice {
		case "", "1":
			// Restore HP and alive status for surviving members
			for _, p := range party {
				p.HP = p.MaxHP
				p.Alive = true
			}
			return party

		case "2":
			if len(party) <= 1 {
				fmt.Println("  Can't remove — you need at least one player.")
				continue
			}
			fmt.Println("\nRemove which player?")
			for i, p := range party {
				fmt.Printf("  %d. %s (%s %s)\n", i+1, p.Name, p.Race, p.Class)
			}
			idx := readInt(reader, "  > ")
			if idx < 1 || idx > len(party) {
				fmt.Println("  Invalid selection.")
				continue
			}
			removed := party[idx-1]
			party = append(party[:idx-1], party[idx:]...)
			fmt.Printf("  Removed %s.\n", removed.Name)

		case "3":
			fmt.Println("\nEdit which player?")
			for i, p := range party {
				fmt.Printf("  %d. %s (%s %s, STR:%d AGI:%d INT:%d CHA:%d)%s\n",
					i+1, p.Name, p.Race, p.Class,
					p.Strength, p.Agility, p.Intelligence, p.Charisma,
					healerTag(p))
			}
			idx := readInt(reader, "  > ")
			if idx < 1 || idx > len(party) {
				fmt.Println("  Invalid selection.")
				continue
			}
			editPlayer(party[idx-1], reader)

		case "4":
			if len(party) >= 6 {
				fmt.Println("  Party is full (max 6).")
				continue
			}
			fmt.Println("\nClasses: fighter, mage, rogue, warrior-mage, mage-rogue, rogue-fighter")
			fmt.Println("Races:   orc, troll, night elf, undead, harpy, goliath")
			p := createPlayer(len(party)+1, reader)
			party = append(party, p)
			fmt.Printf("  → %s added to the party.\n", p.Name)

		case "5":
			return setupParty(reader)

		default:
			fmt.Println("  Enter 1-5.")
		}
	}
}

func editPlayer(p *Character, reader *bufio.Reader) {
	fmt.Printf("\n── Editing %s ──\n", p.Name)
	fmt.Println("  1. Rename")
	fmt.Println("  2. Change class & race (rebuilds stats)")
	fmt.Println("  3. Upgrade a stat (+1)")
	fmt.Println("  4. Toggle healer")
	fmt.Println("  5. Cancel")

	choice := readLine(reader, "  > ")
	switch choice {
	case "1":
		name := sanitize(readLine(reader, fmt.Sprintf("  New name (was %s): ", p.Name)))
		if name != "" {
			p.Name = name
			fmt.Printf("  Renamed to %s.\n", p.Name)
		}
	case "2":
		fmt.Println("  Classes: fighter, mage, rogue, warrior-mage, mage-rogue, rogue-fighter")
		fmt.Println("  Races:   orc, troll, night elf, undead, harpy, goliath")
		var class string
		for {
			class = strings.ToLower(readLine(reader, fmt.Sprintf("  Class (was %s): ", p.Class)))
			if class == "" {
				class = p.Class
			}
			if _, ok := classBase[class]; ok {
				break
			}
			fmt.Println("  No such class. Options: fighter, mage, rogue, warrior-mage, mage-rogue, rogue-fighter")
		}
		var race string
		for {
			race = strings.ToLower(readLine(reader, fmt.Sprintf("  Race (was %s): ", p.Race)))
			if race == "" {
				race = p.Race
			}
			if _, ok := raceBonus[race]; ok {
				break
			}
			fmt.Println("  No such race. Options: orc, troll, night elf, undead, harpy, goliath")
			race = p.Race
		}
		// Rebuild with same name, keeping level and XP
		oldLevel, oldXP, oldName := p.Level, p.XP, p.Name
		rebuilt := newCharacter(oldName, class, race, p.IsHealer)
		rebuilt.Level = oldLevel
		rebuilt.XP = oldXP
		*p = *rebuilt
		fmt.Printf("  → %s rebuilt as %s %s | HP:%d | STR:%d AGI:%d INT:%d CHA:%d\n",
			p.Name, p.Race, p.Class, p.MaxHP, p.Strength, p.Agility, p.Intelligence, p.Charisma)
	case "3":
		p.chooseStat(reader)
	case "4":
		p.IsHealer = !p.IsHealer
		if p.IsHealer {
			fmt.Printf("  %s is now a healer.\n", p.Name)
		} else {
			fmt.Printf("  %s is no longer a healer.\n", p.Name)
		}
	default:
		fmt.Println("  Cancelled.")
	}
}

func createPlayer(num int, reader *bufio.Reader) *Character {
	fmt.Printf("── New Player ──────────────────\n")
	name := sanitize(readLine(reader, fmt.Sprintf("  Name (default: Player%d): ", num)))
	if name == "" {
		name = fmt.Sprintf("Player%d", num)
	}

	var class string
	for {
		class = strings.ToLower(readLine(reader, "  Class (default: fighter): "))
		if class == "" {
			class = "fighter"
		}
		if _, ok := classBase[class]; ok {
			break
		}
		fmt.Println("  No such class. Options: fighter, mage, rogue, warrior-mage, mage-rogue, rogue-fighter")
	}

	suggestedRace := bestRaceForClass(class)
	var race string
	for {
		race = strings.ToLower(readLine(reader, fmt.Sprintf("  Race (default: %s): ", suggestedRace)))
		if race == "" {
			race = suggestedRace
		}
		if _, ok := raceBonus[race]; ok {
			break
		}
		fmt.Println("  No such race. Options: orc, troll, night elf, undead, harpy, goliath")
	}

	isHealer := class == "mage"
	if class == "mage" {
		h := strings.ToLower(readLine(reader, "  Healing mage? (y/n, default: y): "))
		isHealer = h != "n"
	}

	c := newCharacter(name, class, race, isHealer)
	fmt.Printf("  → %s | HP:%d | STR:%d AGI:%d INT:%d CHA:%d | %s (%d dmg)\n",
		c.Name, c.MaxHP, c.Strength, c.Agility, c.Intelligence, c.Charisma,
		c.Weapon.Name, c.Weapon.Damage)
	return c
}

func healerTag(p *Character) string {
	if p.IsHealer {
		return " [healer]"
	}
	return ""
}

// ==================== MAIN ====================

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n╔═══════════════════════════════════╗")
	fmt.Println("║   Micro RPG  Combat Simulator     ║")
	fmt.Println("╚═══════════════════════════════════╝\n")

	party := setupParty(reader)

	for {
		enemies := setupEnemies(reader)

		fmt.Println("\n─────── Combat begins! ───────")
		runCombat(party, enemies, reader)

		cont := strings.ToLower(readLine(reader, "\nContinue fighting? (y/n, default: y): "))
		if cont == "n" {
			fmt.Println("\nThanks for playing. Farewell, adventurers!")
			break
		}

		party = manageParty(party, reader)
	}
}
