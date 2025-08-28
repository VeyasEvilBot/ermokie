package games

import (
	"regexp"
	"strings"
)

type Game struct {
	ID         string
	Name       string
	ShortForms []string
}

type Catalog struct {
	All   []Game
	index map[string]int
}

func NewCatalog() *Catalog {
	c := &Catalog{
		All: []Game{
			{ID: "DemonsSouls", Name: "Demon's Souls", ShortForms: []string{
				"demons souls", "des",
			}},
			{ID: "DarkSouls1", Name: "Dark Souls 1", ShortForms: []string{
				"ds1", "dark souls remastered", "dsr", "dark souls",
				"prepare to die", "ptde", "prepairtodie", // common typo in samples
			}},
			{ID: "DarkSouls2", Name: "Dark Souls 2", ShortForms: []string{
				"ds2", "scholar", "sotfs", "sotf", "ds2 sotf", "scholar of the first sin",
			}},
			{ID: "DarkSouls3", Name: "Dark Souls 3", ShortForms: []string{
				"ds3", "ashesofariandel", "the ringed city", "theringedcity", "ringed city",
			}},
			{ID: "Bloodborne", Name: "Bloodborne", ShortForms: []string{
				"bb", "the old hunters", "theoldhunters", "old hunters", "oldhunters",
			}},
			{ID: "Sekiro", Name: "Sekiro", ShortForms: []string{
				"sekiro shadows die twice", "sk", "ssdt",
			}},
			{ID: "EldenRing", Name: "Elden Ring", ShortForms: []string{"er"}},
			{ID: "MarathonSlop", Name: "Marathon Slop", ShortForms: []string{"marathon slop"}},
			{ID: "ResidentEvil0", Name: "Resident Evil 0", ShortForms: []string{"re0", "zero"}},
			{ID: "ResidentEvil1", Name: "Resident Evil", ShortForms: []string{"re1", "biohazard"}},
			{ID: "ResidentEvil2", Name: "Resident Evil 2", ShortForms: []string{"re2"}},
			{ID: "ResidentEvil3", Name: "Resident Evil 3", ShortForms: []string{"re3"}},
			{ID: "ResidentEvil4", Name: "Resident Evil 4", ShortForms: []string{"re4"}},
			{ID: "ResidentEvil7", Name: "Resident Evil 7", ShortForms: []string{"re7", "resident evil biohazard"}},
			{ID: "ResidentEvil8", Name: "Resident Evil 8", ShortForms: []string{"re8", "village", "resident evil village"}},
			{ID: "ResidentEvilSpinoffs", Name: "Resident Evil Spinoffs", ShortForms: []string{"re spinoffs", "spinoffs"}},
			{ID: "Hades", Name: "Hades", ShortForms: []string{}},
			{ID: "Cuphead", Name: "Cuphead", ShortForms: []string{}},
			{ID: "HollowKnight", Name: "Hollow Knight", ShortForms: []string{"hk"}},
			{ID: "HollowKnightSilksong", Name: "Hollow Knight: Silksong", ShortForms: []string{"silksong", "hks", "hk2"}},
			{ID: "Celeste", Name: "Celeste", ShortForms: []string{}},
			{ID: "CrashBandicoot1", Name: "Crash Bandicoot", ShortForms: []string{"cb1", "crash 1"}},
			{ID: "CrashBandicoot2CortexStrikesBack", Name: "Crash Bandicoot 2: Cortex Strikes Back", ShortForms: []string{"cb2", "crash 2"}},
			{ID: "CrashBandicoot3Warped", Name: "Crash Bandicoot 3: Warped", ShortForms: []string{"cb3", "crash 3", "warped"}},
			{ID: "CrashBandicoot4ItsAboutTime", Name: "Crash Bandicoot 4: It’s About Time", ShortForms: []string{"cb4", "crash 4", "its about time"}},
			{ID: "OcarinaOfTime", Name: "Ocarina of Time", ShortForms: []string{"oot", "zelda oot"}},
			{ID: "MajorasMask", Name: "Majora's Mask", ShortForms: []string{"mm", "zelda mm"}},
			{ID: "BreathOfTheWild", Name: "Breath of the Wild", ShortForms: []string{"botw"}},
			{ID: "TearsOfTheKingdom", Name: "Tears of the Kingdom", ShortForms: []string{"totk"}},
			{ID: "Blasphemous", Name: "Blasphemous", ShortForms: []string{}},
			{ID: "Blasphemous2", Name: "Blasphemous 2", ShortForms: []string{"blasphemous ii"}},
			{ID: "SilentHill1", Name: "Silent Hill", ShortForms: []string{"sh1"}},
			{ID: "SilentHill2Remake", Name: "Silent Hill 2 Remake", ShortForms: []string{"sh2r"}},
			{ID: "SilentHill2Classic", Name: "Silent Hill 2 Classic", ShortForms: []string{"sh2"}},
			{ID: "SilentHill3", Name: "Silent Hill 3", ShortForms: []string{"sh3"}},
			{ID: "SilentHill4", Name: "Silent Hill 4", ShortForms: []string{"sh4", "the room"}},
			{ID: "SilentHillOrigins", Name: "Silent Hill Origins", ShortForms: []string{"sho", "origins"}},
			{ID: "Dishonored1", Name: "Dishonored", ShortForms: []string{"d1"}},
			{ID: "Dishonored2", Name: "Dishonored 2", ShortForms: []string{"d2"}},
			{ID: "DishonoredDeathOfTheOutsider", Name: "Dishonored: Death of the Outsider", ShortForms: []string{"doto", "dishonored dto"}},
			{ID: "TormentedSouls", Name: "Tormented Souls", ShortForms: []string{}},
			{ID: "LiesOfP", Name: "Lies of P", ShortForms: []string{"lop"}},
			{ID: "OblivionClassic", Name: "Oblivion Classic", ShortForms: []string{"oblivion"}},
			{ID: "OblivionRemastered", Name: "Oblivion Remastered", ShortForms: []string{"oblivion remaster", "oblivion remastered"}},
			{ID: "Skyrim", Name: "Skyrim", ShortForms: []string{"tes5", "the elder scrolls v"}},
			{ID: "Minecraft", Name: "Minecraft", ShortForms: []string{"mc"}},
			{ID: "ArkhamAsylum", Name: "Arkham Asylum", ShortForms: []string{"baa"}},
			{ID: "ArkhamOrigins", Name: "Arkham Origins", ShortForms: []string{"bao"}},
			{ID: "ArkhamKnight", Name: "Arkham Knight", ShortForms: []string{"bak"}},
			{ID: "DeadCells", Name: "Dead Cells", ShortForms: []string{}},
			{ID: "Fallout3", Name: "Fallout 3", ShortForms: []string{"fo3"}},
			{ID: "FalloutNewVegas", Name: "Fallout: New Vegas", ShortForms: []string{"fnv", "new vegas"}},
			{ID: "Fallout4", Name: "Fallout 4", ShortForms: []string{"fo4"}},
			{ID: "TheBindingOfIsaac", Name: "The Binding of Isaac", ShortForms: []string{"tboi", "boi"}},
			{ID: "OriAndTheBlindForest", Name: "Ori and the Blind Forest", ShortForms: []string{"bf"}},
			{ID: "OriAndTheWillOfTheWisps", Name: "Ori and the Will of the Wisps", ShortForms: []string{"wotw", "will of the wisps"}},
			{ID: "ClairObscurExpedition33", Name: "Clair Obscur: Expedition 33", ShortForms: []string{"expedition 33", "co:e33"}},
			{ID: "Thymesia", Name: "Thymesia", ShortForms: []string{}},
			{ID: "SuperMario64", Name: "Super Mario 64", ShortForms: []string{"sm64"}},
			{ID: "SuperMarioOdyssey", Name: "Super Mario Odyssey", ShortForms: []string{"smo"}},
		},
	}
	c.buildIndex()
	return c
}

func (c *Catalog) buildIndex() {
	c.index = make(map[string]int, len(c.All)*3)

	put := func(key string, idx int) {
		key = strings.ToLower(strings.TrimSpace(key))
		if key == "" {
			return
		}
		if _, exists := c.index[key]; !exists {
			c.index[key] = idx
		}
	}

	for i, g := range c.All {
		put(g.Name, i)
		put(g.ID, i)
		for _, s := range g.ShortForms {
			put(s, i)
		}
	}
}

func (c *Catalog) Parse(s string) (Game, bool) {
	if c.index == nil {
		c.buildIndex()
	}
	key := strings.ToLower(strings.TrimSpace(s))
	if key == "" {
		return Game{}, false
	}
	if i, ok := c.index[key]; ok && i >= 0 && i < len(c.All) {
		return c.All[i], true
	}
	return Game{}, false
}

func (c *Catalog) MustParse(s string) Game {
	g, _ := c.Parse(s)
	return g
}

var (
	bracketOrParen = regexp.MustCompile(`[([]`)
	langTag        = regexp.MustCompile(`\[[A-Za-z]{2,5}(?:[-/][A-Za-z0-9]{2,5})?\]$`)
	spaceCollapse  = regexp.MustCompile(`\s+`)
)

func ExtractGameToken(profile string) string {
	s := strings.TrimSpace(profile)
	s = langTag.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	idx := len(s)
	if loc := bracketOrParen.FindStringIndex(s); loc != nil {
		idx = loc[0]
	}
	s = strings.TrimSpace(s[:idx])
	s = spaceCollapse.ReplaceAllString(s, " ")
	return s
}

func ParseProfileGame(cat *Catalog, profile string) (Game, bool) {
	token := ExtractGameToken(profile)

	if g, ok := cat.Parse(token); ok {
		return g, true
	}
	if strings.Contains(token, " ") {
		if g, ok := cat.Parse(strings.ReplaceAll(token, " ", "")); ok {
			return g, true
		}
	}

	words := splitCamelOrAlphaRuns(token)
	for i := len(words); i >= 1; i-- {
		prefix := strings.Join(words[:i], "")
		if g, ok := cat.Parse(prefix); ok {
			return g, true
		}
		prefixSpaced := strings.Join(words[:i], " ")
		if g, ok := cat.Parse(prefixSpaced); ok {
			return g, true
		}
	}

	s := strings.ToLower(profile)
	for _, game := range cat.All {
		cands := make([]string, 0, 2+len(game.ShortForms))
		cands = append(cands, strings.ToLower(game.Name))
		cands = append(cands, strings.ToLower(game.ID))
		for _, a := range game.ShortForms {
			cands = append(cands, strings.ToLower(a))
		}
		for _, cand := range cands {
			if cand != "" && strings.Contains(s, cand) {
				return game, true
			}
		}
	}

	return Game{}, false
}

func splitCamelOrAlphaRuns(s string) []string {
	if s == "" {
		return nil
	}
	var b strings.Builder
	b.Grow(len(s) * 2)
	for i, r := range s {
		if i > 0 {
			prev := rune(s[i-1])
			if (isLower(prev) && isUpper(r)) ||
				(isDigit(prev) && isUpper(r)) ||
				(isLower(prev) && isDigit(r)) ||
				(isUpper(prev) && isDigit(r)) {
				b.WriteByte(' ')
			}
		}
		b.WriteRune(r)
	}
	return strings.Fields(b.String())
}

func isLower(r rune) bool { return r >= 'a' && r <= 'z' }
func isUpper(r rune) bool { return r >= 'A' && r <= 'Z' }
func isDigit(r rune) bool { return r >= '0' && r <= '9' }
