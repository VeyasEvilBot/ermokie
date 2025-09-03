package categories

import (
	"regexp"
	"strings"
)

type Category struct {
	ID         string
	Name       string
	ShortForms []string
}

type Catalog struct {
	All   []Category
	index map[string]int
}

func NewCatalog() *Catalog {
	c := &Catalog{
		All: []Category{
			{ID: "AnyPercent", Name: "Any%", ShortForms: []string{"any", "any%", "any percent"}},
			{ID: "AllBosses", Name: "All Bosses", ShortForms: []string{"allbosses", "ab"}},
			{ID: "AllGreatRunes", Name: "All Great Runes", ShortForms: []string{"allgreatrunes", "agr"}},
			{ID: "AllAchievements", Name: "All Achievements", ShortForms: []string{
				"alla", "achievements", "all cheevos", "cheevos", "100%",
			}},
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

	for i, cat := range c.All {
		put(cat.Name, i)
		put(cat.ID, i)
		for _, s := range cat.ShortForms {
			put(s, i)
		}
	}
}

var percentCategory = regexp.MustCompile(`^\d+%$`)

func (c *Catalog) Parse(s string) (Category, bool) {
	if c.index == nil {
		c.buildIndex()
	}
	key := strings.ToLower(strings.TrimSpace(s))
	if key == "" {
		return Category{}, false
	}

	if percentCategory.MatchString(key) {
		for _, cat := range c.All {
			if cat.ID == "AllAchievements" {
				return cat, true
			}
		}
	}

	if i, ok := c.index[key]; ok && i >= 0 && i < len(c.All) {
		return c.All[i], true
	}
	return Category{}, false
}

func (c *Catalog) MustParse(s string) Category {
	cat, _ := c.Parse(s)
	return cat
}

var (
	bracketOrParen = regexp.MustCompile(`[([]`)
	langTag        = regexp.MustCompile(`\[[A-Za-z]{2,5}(?:[-/][A-Za-z0-9]{2,5})?\]$`)
)

func ExtractCategoryToken(profile string) string {
	s := strings.TrimSpace(profile)
	s = langTag.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)

	if loc := bracketOrParen.FindStringIndex(s); loc != nil {
		rest := s[loc[0]:]
		end := strings.IndexAny(rest, ")]")
		if end > 0 {
			token := rest[1:end]
			return strings.TrimSpace(token)
		}
	}
	return ""
}

func ParseProfileCategory(cat *Catalog, profile string) (Category, bool) {
	token := ExtractCategoryToken(profile)
	if token == "" {
		return Category{}, false
	}
	return cat.Parse(token)
}
